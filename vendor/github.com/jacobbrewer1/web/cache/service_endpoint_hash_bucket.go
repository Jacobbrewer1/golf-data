package cache

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
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

const (
	// k8sNameLabel is the label used to identify the Kubernetes resource name.
	k8sNameLabel = "app.kubernetes.io/name"
)

// Ensures that ServiceEndpointHashBucket implements the HashBucket interface.
var _ HashBucket = (*ServiceEndpointHashBucket)(nil)

// ServiceEndpointHashBucket represents a mechanism which determines whether the current application instance should process
// a particular key. The bucket size is determined by the number of active endpoints in the supplied Kubernetes service.
type ServiceEndpointHashBucket struct {
	// mut is a read-write mutex used to ensure thread-safe access to the hash ring and other shared resources.
	mut *sync.RWMutex

	// startOnce is a sync.Once instance used to ensure that the Start method is only called once.
	startOnce sync.Once

	// startErr is an error that captures any errors encountered during the Start process.
	startErr error

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
func (sb *ServiceEndpointHashBucket) Start(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context cannot be nil")
	}

	sb.startOnce.Do(func() {
		// Get the initial list of endpoint slices for the application
		endpointSliceList, err := sb.kubeClient.DiscoveryV1().EndpointSlices(sb.appNamespace).List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", k8sNameLabel, sb.appName),
		})
		if err != nil {
			sb.startErr = fmt.Errorf("error listing endpoint slices: %w", err)
			return
		} else if len(endpointSliceList.Items) == 0 {
			sb.startErr = fmt.Errorf("no endpoint slices found for application %s in namespace %s", sb.appName, sb.appNamespace)
			return
		}

		// Combine all endpoints from all slices into a single set
		currentHostSet := slices.NewSet[string]()
		for i := range endpointSliceList.Items {
			endpointSlice := &endpointSliceList.Items[i]
			sliceHosts := endpointSliceToSet(endpointSlice)
			sliceHosts.Each(func(host string) {
				currentHostSet.Add(host)
			})
		}

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
		if _, err := sb.endpointsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{ // nolint:revive // Traditional error handling
			UpdateFunc: sb.onEndpointUpdate,
		}); err != nil {
			sb.startErr = fmt.Errorf("error adding event handler to endpoints informer: %w", err)
			return
		}
	})
	return sb.startErr
}

// InBucket checks if the given key is assigned to the current application instance
// based on the consistent hash ring.
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

	// Check if the updated endpoint slice matches the application
	if coreNewEndpoints.GetNamespace() != sb.appNamespace {
		return
	}

	// Check if this EndpointSlice belongs to our application
	generateName := coreNewEndpoints.GetGenerateName()
	if coreNewEndpoints.Labels[k8sNameLabel] != sb.appName && (generateName == "" || !strings.HasPrefix(generateName, sb.appName)) {
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

// Shutdown cleans up the resources used by the ServiceEndpointHashBucket.
func (sb *ServiceEndpointHashBucket) Shutdown() {
	sb.informerFactory.Shutdown()
}

// endpointSliceToSet converts a Kubernetes EndpointSlice into a set of pod names.
func endpointSliceToSet(endpointSlice *discoveryv1.EndpointSlice) *slices.Set[string] {
	s := slices.NewSet[string]()
	for _, endpoint := range endpointSlice.Endpoints {
		if endpoint.TargetRef != nil && endpoint.TargetRef.Kind == "Pod" {
			s.Add(endpoint.TargetRef.Name)
		}
	}
	return s
}
