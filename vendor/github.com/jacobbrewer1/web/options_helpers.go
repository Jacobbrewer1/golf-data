package web

import (
	"time"

	"k8s.io/client-go/informers"
)

// initKubernetesInformerFactory initializes the Kubernetes SharedInformerFactory if it is not already initialized.
//
// This function sets up a SharedInformerFactory for the application using the provided Kubernetes client
// and informer options.
func initKubernetesInformerFactory(a *App, informerOptions ...informers.SharedInformerOption) {
	// Set up an informer factory if one does not exist.
	if a.kubernetesInformerFactory != nil {
		return
	}

	// Set up a factory and informer to keep track of Kubernetes objects.
	a.kubernetesInformerFactory = informers.NewSharedInformerFactoryWithOptions(a.KubeClient(), time.Second*30, informerOptions...)
}
