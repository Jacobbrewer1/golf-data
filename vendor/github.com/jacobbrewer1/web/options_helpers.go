package web

import (
	"time"

	"k8s.io/client-go/informers"
)

// initKubernetesInformerFactory initialises the kubernetes informer factory for App.
func initKubernetesInformerFactory(a *App, options ...informers.SharedInformerOption) {
	// Set up an informer factory if one does not exist.
	if a.kubernetesInformerFactory != nil {
		return
	}

	kubeClient := a.KubeClient()

	// Set up a factory and informer to keep track of Kubernetes objects.
	a.kubernetesInformerFactory = informers.NewSharedInformerFactoryWithOptions(kubeClient, time.Second*30, options...)
}
