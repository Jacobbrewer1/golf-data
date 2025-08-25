package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// UpsertableResource defines an interface for Kubernetes resources that can be upserted.
type UpsertableResource interface {
	// GetName retrieves the name of the resource.
	GetName() string

	// GetUID retrieves the unique identifier (UID) of the resource.
	GetUID() types.UID

	// SetUID sets the unique identifier (UID) of the resource.
	SetUID(types.UID)

	// GetResourceVersion retrieves the resource version, which is used
	// for optimistic concurrency control in Kubernetes.
	GetResourceVersion() string

	// SetResourceVersion sets the resource version of the resource.
	SetResourceVersion(string)
}

// Upserter defines a generic interface for managing Kubernetes resources.
type Upserter[T UpsertableResource] interface {
	// Create creates a new resource in the Kubernetes cluster.
	Create(ctx context.Context, obj T, opts metav1.CreateOptions) (T, error)

	// Update updates an existing resource in the Kubernetes cluster.
	Update(ctx context.Context, obj T, opts metav1.UpdateOptions) (T, error)

	// Get retrieves a resource from the Kubernetes cluster by its name.
	Get(ctx context.Context, name string, opts metav1.GetOptions) (T, error)
}

// UpsertResource attempts to create or update a Kubernetes resource.
func UpsertResource(ctx context.Context, kubeClient kubernetes.Interface, resource metav1.Object) error {
	switch v := resource.(type) {
	case *corev1.ConfigMap:
		return upsert(ctx, kubeClient.CoreV1().ConfigMaps(resource.GetNamespace()), v)
	case *corev1.Secret:
		return upsert(ctx, kubeClient.CoreV1().Secrets(resource.GetNamespace()), v)
	default:
		return fmt.Errorf("unsupported resource type: %T", resource)
	}
}

// upsert performs an upsert operation (create or update) for a Kubernetes resource.
func upsert[U Upserter[T], T UpsertableResource](ctx context.Context, upserter U, obj T) error {
	_, err := upserter.Get(ctx, obj.GetName(), metav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return fmt.Errorf("failed to get %T: %w", obj, err)
		}

		obj.SetResourceVersion("") // Reset resource version for creation
		if _, err := upserter.Create(ctx, obj, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("failed to create %T: %w", obj, err)
		}
		return nil
	}

	obj, err = upserter.Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update %T: %w", obj, err)
	}

	return nil
}
