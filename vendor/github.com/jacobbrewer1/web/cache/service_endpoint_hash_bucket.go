package cache

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jacobbrewer1/web/logging"
	"github.com/jacobbrewer1/web/slices"
	"github.com/serialx/hashring"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

var _ HashBucket = new(ServiceEndpointHashBucket)

// ServiceEndpointHashBucket represents a mechanism which determines whether the current application instance should process
// a particular key. The bucket size is determined by the number of active endpoints in the supplied Kubernetes service.
type ServiceEndpointHashBucket struct {
	mut               sync.Mutex
	hr                *hashring.HashRing
	l                 *slog.Logger
	k                 kubernetes.Interface
	appName           string
	appNamespace      string
	thisPod           string
	informerFactory   informers.SharedInformerFactory
	endpointsInformer cache.SharedIndexInformer
}

// NewServiceEndpointHashBucket returns a new serviceEndpointHashBucket.
func NewServiceEndpointHashBucket(
	l *slog.Logger,
	kubeClient kubernetes.Interface,
	serviceName string,
	serviceNamespace string,
	thisPod string,
) *ServiceEndpointHashBucket {
	informerFactory := informers.NewSharedInformerFactory(kubeClient, 10*time.Second)
	endpointsInformer := informerFactory.Core().V1().Endpoints().Informer()
	return &ServiceEndpointHashBucket{
		l:                 l,
		k:                 kubeClient,
		appName:           serviceName,
		appNamespace:      serviceNamespace,
		thisPod:           thisPod,
		informerFactory:   informerFactory,
		endpointsInformer: endpointsInformer,
	}
}

// Start starts the hash bucket processing.
func (sb *ServiceEndpointHashBucket) Start(ctx context.Context) error {
	currentEndpoints, err := sb.k.CoreV1().Endpoints(sb.appNamespace).Get(ctx, sb.appName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting initial endpoints: %w", err)
	}

	currentHostSet := endpointsToSet(currentEndpoints)
	sb.l.Info("initialising hash ring with hosts", slog.Any(logging.KeyHosts, currentHostSet.Items()))

	sb.mut.Lock()
	defer sb.mut.Unlock()
	sb.hr = hashring.New(currentHostSet.Items())

	sb.informerFactory.Start(ctx.Done())
	sb.informerFactory.WaitForCacheSync(ctx.Done())

	_, err = sb.endpointsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: sb.onEndpointUpdate,
	})
	return err
}

// InBucket returns whether the supplied key is processed by this endpoint.
func (sb *ServiceEndpointHashBucket) InBucket(key string) bool {
	sb.mut.Lock()
	defer sb.mut.Unlock()

	node, _ := sb.hr.GetNode(key)
	return node == sb.thisPod
}

// onEndpointUpdate is called when a change to an endpoint is made. If the endpoint that has changed is
// related to this service, the internal hashring is updated to represent the new set of nodes.
func (sb *ServiceEndpointHashBucket) onEndpointUpdate(oldEndpoints, newEndpoints any) {
	coreOldEndpoints, ok := oldEndpoints.(*corev1.Endpoints)
	if !ok {
		return
	}

	coreNewEndpoints, ok := newEndpoints.(*corev1.Endpoints)
	if !ok {
		return
	}

	if coreNewEndpoints.Name != sb.appName || coreNewEndpoints.Namespace != sb.appNamespace {
		return
	}

	a := endpointsToSet(coreOldEndpoints)
	b := endpointsToSet(coreNewEndpoints)
	removed := a.Difference(b)
	added := b.Difference(a)

	sb.mut.Lock()
	defer sb.mut.Unlock()

	if len(added.Items()) == 0 && len(removed.Items()) == 0 {
		return
	}

	removed.Each(func(item string) {
		sb.l.Info("removing node from hashring", slog.String(logging.KeyItem, item))
		sb.hr = sb.hr.RemoveNode(item)
	})

	added.Each(func(item string) {
		sb.l.Info("adding node to hashring", slog.String(logging.KeyItem, item))
		sb.hr = sb.hr.AddNode(item)
	})

	nodes, _ := sb.hr.GetNodes("", sb.hr.Size())
	sb.l.Info("hashring state", slog.Any(logging.KeyState, nodes))
}

// endpointsToSet converts input endpoints into a Set of IP addresses.
func endpointsToSet(endpoints *corev1.Endpoints) *slices.Set[string] {
	s := slices.NewSet[string]()

	for _, subset := range endpoints.Subsets {
		for _, address := range subset.Addresses {
			if address.TargetRef != nil {
				s.Add(address.TargetRef.Name)
			}
		}
	}

	return s
}

// Shutdown stops the informer factory and cleans up resources.
func (sb *ServiceEndpointHashBucket) Shutdown() {
	sb.informerFactory.Shutdown()
}
