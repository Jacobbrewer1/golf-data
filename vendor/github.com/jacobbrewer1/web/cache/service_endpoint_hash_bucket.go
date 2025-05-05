package cache

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/serialx/hashring"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/jacobbrewer1/web/logging"
	"github.com/jacobbrewer1/web/slices"
)

// Ensures that ServiceEndpointHashBucket implements the HashBucket interface.
//
// This line is a compile-time check to verify that the ServiceEndpointHashBucket
// struct satisfies all the methods defined in the HashBucket interface.
var _ HashBucket = new(ServiceEndpointHashBucket)

// ServiceEndpointHashBucket represents a mechanism which determines whether the current application instance should process
// a particular key. The bucket size is determined by the number of active endpoints in the supplied Kubernetes service.
//
// ServiceEndpointHashBucket maintains a consistent hashing ring where each node corresponds to a Kubernetes pod
// running this application. It dynamically updates the hash ring based on endpoint changes detected through
// the Kubernetes API, ensuring keys are consistently distributed across available application instances.
type ServiceEndpointHashBucket struct {
	// mut is a read-write mutex used to ensure thread-safe access to the hash ring and other shared resources.
	mut *sync.RWMutex

	// hr represents the consistent hash ring used to distribute keys among application instances.
	hr *hashring.HashRing

	// l is the logger used for logging events and errors.
	l *slog.Logger

	// kubeClient is the Kubernetes client interface used to interact with the Kubernetes API.
	kubeClient kubernetes.Interface

	// appName is the name of the application associated with this bucket.
	appName string

	// appNamespace is the namespace of the application in the Kubernetes cluster.
	appNamespace string

	// thisPod is the name of the current pod running the application.
	thisPod string

	// informerFactory is the shared informer factory used to manage informers for Kubernetes resources.
	informerFactory informers.SharedInformerFactory

	// endpointsInformer is the shared index informer used to monitor endpoint slices in the Kubernetes cluster.
	endpointsInformer cache.SharedIndexInformer
}

// NewServiceEndpointHashBucket initializes and returns a new instance of ServiceEndpointHashBucket.
//
// Parameters:
//   - l: A logger instance used for logging events and errors.
//   - kubeClient: A Kubernetes client interface for interacting with the Kubernetes API.
//   - appName: The name of the application associated with this bucket.
//   - appNamespace: The namespace of the application in the Kubernetes cluster.
//   - thisPod: The name of the current pod running the application.
//
// Returns:
//   - A pointer to a newly created ServiceEndpointHashBucket instance, which includes
//     a consistent hash ring and Kubernetes informers for monitoring endpoint slices.
func NewServiceEndpointHashBucket(
	l *slog.Logger,
	kubeClient kubernetes.Interface,
	appName, appNamespace, thisPod string,
) *ServiceEndpointHashBucket {
	informerFactory := informers.NewSharedInformerFactory(kubeClient, 10*time.Second)
	endpointsInformer := informerFactory.Discovery().V1().EndpointSlices().Informer()
	return &ServiceEndpointHashBucket{
		mut:               new(sync.RWMutex),
		l:                 l,
		kubeClient:        kubeClient,
		appName:           appName,
		appNamespace:      appNamespace,
		thisPod:           thisPod,
		informerFactory:   informerFactory,
		endpointsInformer: endpointsInformer,
	}
}

// Start initializes and starts the hash bucket processing by setting up the hash ring
// and Kubernetes informers to monitor endpoint changes.
//
// Parameters:
//   - ctx: The context used to manage the lifecycle of the operation.
//
// Steps:
//  1. Retrieves the initial list of endpoint slices for the application from the Kubernetes API.
//  2. Converts the endpoint slice into a set of hostnames and initializes the hash ring with these hosts.
//  3. Starts the shared informer factory to monitor changes in endpoint slices.
//  4. Waits for the informer cache to synchronize.
//  5. Adds an event handler to the endpoints informer to handle updates to endpoint slices.
//
// Returns:
//   - An error if there is an issue retrieving the initial endpoints or adding the event handler.
func (sb *ServiceEndpointHashBucket) Start(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context cannot be nil")
	}

	// Get the initial list of endpoint slices for the application
	endpointSliceList, err := sb.kubeClient.DiscoveryV1().EndpointSlices(sb.appNamespace).Get(ctx, sb.appName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting initial endpoints: %w", err)
	}

	// Convert the endpoint slice into a set of hostnames
	currentHostSet := endpointSliceToSet(endpointSliceList)
	sb.l.Info("initialising hash ring with hosts", slog.Any(logging.KeyHosts, currentHostSet.Items()))

	// Lock the mutex to ensure thread-safe access to the hash ring
	sb.mut.Lock()
	defer sb.mut.Unlock()

	// Initialize the hash ring with the current set of hosts
	sb.hr = hashring.New(currentHostSet.Items())

	// Start the shared informer factory to monitor changes in endpoint slices and wait for cache sync
	sb.informerFactory.Start(ctx.Done())
	sb.informerFactory.WaitForCacheSync(ctx.Done())

	// Add an event handler to the endpoints informer to handle updates to endpoint slices
	if _, err := sb.endpointsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: sb.onEndpointUpdate,
	}); err != nil {
		return fmt.Errorf("error adding event handler to endpoints informer: %w", err)
	}
	return nil
}

// InBucket checks if the given key is assigned to the current application instance
// based on the consistent hash ring.
//
// Parameters:
//   - key: The key to check against the hash ring.
//
// Returns:
//   - true if the key is assigned to the current pod (thisPod), false otherwise.
//
// Notes:
//   - The method acquires a read lock to ensure thread-safe access to the hash ring.
//   - If the hash ring is not initialized, an error is logged and the method returns false.
func (sb *ServiceEndpointHashBucket) InBucket(key string) bool {
	sb.mut.RLock()
	defer sb.mut.RUnlock()

	if sb.hr == nil {
		sb.l.Error("hash ring is not initialized - has Start() been called?")
		return false
	}

	node, _ := sb.hr.GetNode(key)
	return node == sb.thisPod
}

// onEndpointUpdate handles updates to endpoint slices in the Kubernetes cluster.
//
// Parameters:
//   - oldEndpoints: The previous state of the endpoint slice, expected to be of type *discoveryv1.EndpointSlice.
//   - newEndpoints: The updated state of the endpoint slice, expected to be of type *discoveryv1.EndpointSlice.
//
// Behavior:
//   - If the provided endpoints are not of the expected type, the function returns immediately.
//   - If the updated endpoint slice does not match the application name or namespace, the function returns.
//   - Computes the difference between the old and new endpoint slices to determine added and removed nodes.
//   - Updates the internal consistent hash ring by removing nodes that are no longer present and adding new nodes.
//   - Logs the changes to the hash ring and its current state.
//
// Notes:
//   - The function acquires a write lock to ensure thread-safe updates to the hash ring.
func (sb *ServiceEndpointHashBucket) onEndpointUpdate(oldEndpoints, newEndpoints any) {
	// Confirm type assertions for old endpoints
	coreOldEndpoints, ok := oldEndpoints.(*discoveryv1.EndpointSlice)
	if !ok {
		return
	}

	// Confirm type assertions for new endpoints
	coreNewEndpoints, ok := newEndpoints.(*discoveryv1.EndpointSlice)
	if !ok {
		return
	}

	// Check if the updated endpoint slice matches the application name and namespace
	if coreNewEndpoints.Name != sb.appName || coreNewEndpoints.Namespace != sb.appNamespace {
		return
	}

	// Convert the old and new endpoint slices into sets of pod names
	a := endpointSliceToSet(coreOldEndpoints)
	b := endpointSliceToSet(coreNewEndpoints)

	// Compute the difference between the old and new sets to find added and removed nodes
	removed := a.Difference(b)
	added := b.Difference(a)

	// Do we need to continue? Exit early if there are no changes
	if len(added.Items()) == 0 && len(removed.Items()) == 0 {
		return
	}

	// Lock the mutex to ensure thread-safe updates to the hash ring
	sb.mut.Lock()
	defer sb.mut.Unlock()

	// Update the hash ring by removing and adding nodes based on the computed differences
	removed.Each(func(item string) {
		sb.l.Info("removing node from hashring", slog.String(logging.KeyItem, item))
		sb.hr = sb.hr.RemoveNode(item)
	})

	// Add new nodes to the hash ring
	added.Each(func(item string) {
		sb.l.Info("adding node to hashring", slog.String(logging.KeyItem, item))
		sb.hr = sb.hr.AddNode(item)
	})

	// Log the current state of the hash ring
	nodes, _ := sb.hr.GetNodes("", sb.hr.Size())
	sb.l.Info("hashring state", slog.Any(logging.KeyState, nodes))
}

// Shutdown stops the informer factory and cleans up resources.
//
// Behavior:
//   - Calls the Shutdown method on the shared informer factory to stop all informers
//     and release associated resources.
//
// Notes:
//   - This method should be called during application shutdown to ensure proper cleanup
//     of Kubernetes informers and avoid resource leaks.
func (sb *ServiceEndpointHashBucket) Shutdown() {
	sb.informerFactory.Shutdown()
}

// endpointSliceToSet converts a Kubernetes EndpointSlice into a set of pod names.
//
// Parameters:
//   - endpointSlice: A pointer to a discoveryv1.EndpointSlice object containing endpoint information.
//
// Returns:
//   - A pointer to a slices.Set[string] containing the names of pods extracted from the EndpointSlice.
//
// Notes:
//   - Only endpoints with a non-nil TargetRef and a TargetRef.Kind of "Pod" are included in the set.
func endpointSliceToSet(endpointSlice *discoveryv1.EndpointSlice) *slices.Set[string] {
	s := slices.NewSet[string]()
	for _, endpoint := range endpointSlice.Endpoints {
		if endpoint.TargetRef != nil && endpoint.TargetRef.Kind == "Pod" {
			s.Add(endpoint.TargetRef.Name)
		}
	}
	return s
}
