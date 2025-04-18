package web

import (
	"time"

	"k8s.io/client-go/informers"
)

// initKubernetesInformerFactory initializes the Kubernetes SharedInformerFactory if not already initialized.
// It sets up the factory with the given informer options and the application's Kubernetes client.
func initKubernetesInformerFactory(a *App, informerOptions ...informers.SharedInformerOption) {
	// Set up an informer factory if one does not exist.
	if a.kubernetesInformerFactory != nil {
		return
	}

	// Set up a factory and informer to keep track of Kubernetes objects.
	a.kubernetesInformerFactory = informers.NewSharedInformerFactoryWithOptions(a.KubeClient(), time.Second*30, informerOptions...)
}
