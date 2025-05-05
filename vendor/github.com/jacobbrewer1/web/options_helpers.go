package web

import (
	"time"

	"k8s.io/client-go/informers"
)

// initKubernetesInformerFactory initializes the Kubernetes SharedInformerFactory if it is not already initialized.
//
// This function sets up a SharedInformerFactory for the application using the provided Kubernetes client
// and informer options. The factory is used to manage and cache Kubernetes objects.
//
// Parameters:
//   - a: A pointer to the App struct, which contains the application's Kubernetes client and other configurations.
//   - informerOptions: A variadic list of SharedInformerOption values to configure the SharedInformerFactory.
//
// Behavior:
//   - Checks if the application's Kubernetes SharedInformerFactory is already initialized.
//   - If not initialized, creates a new SharedInformerFactory with the provided Kubernetes client and options.
//   - The factory is configured to resync every 30 seconds.
//
// Notes:
//   - This function does nothing if the SharedInformerFactory is already initialized.
func initKubernetesInformerFactory(a *App, informerOptions ...informers.SharedInformerOption) {
	// Set up an informer factory if one does not exist.
	if a.kubernetesInformerFactory != nil {
		return
	}

	// Set up a factory and informer to keep track of Kubernetes objects.
	a.kubernetesInformerFactory = informers.NewSharedInformerFactoryWithOptions(a.KubeClient(), time.Second*30, informerOptions...)
}
