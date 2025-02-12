package k8s

import (
	"context"
	"fmt"
	"log"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createNamespace(namespace string) (string, error) {
	ctx := context.TODO()
	clientset, _ := K8sConfig()
	_, err := clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil {
		log.Printf("Namespace %s already exists\n", namespace)

		return fmt.Sprintf("Namespace %s already exists\n", namespace), nil
	}

	// If not found, create the namespace
	fmt.Printf("Creating namespace: %s\n", namespace)
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err = clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		log.Println("Failed to create namespace", err)
		return fmt.Sprintf("Failed to create namespace", err), fmt.Errorf("failed to create namespace: %w", err)
	}

	log.Printf("Namespace %s created successfully\n", namespace)
	return fmt.Sprintf("Namespace %s created successfully\n", namespace), nil
}
