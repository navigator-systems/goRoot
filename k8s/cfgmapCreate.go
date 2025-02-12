package k8s

import (
	"context"
	"log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createUpdateConfigMap(name, data, namespace string) (string, error) {

	clientset, _ := K8sConfig()
	// Check if configMap is already created
	ctx := context.TODO()

	_, err := clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})

	if err == nil {
		log.Println("ConfigMap already exists, updating...")
		_, err = clientset.CoreV1().ConfigMaps(namespace).Update(ctx, &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Data:       map[string]string{name: data},
		}, metav1.UpdateOptions{})
	} else {
		log.Println("Creating new ConfigMap...")
		_, err = clientset.CoreV1().ConfigMaps(namespace).Create(ctx, &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Data:       map[string]string{name: data},
		}, metav1.CreateOptions{})
	}
	if err != nil {

		log.Println("Failed to create/update ConfigMap:", err)
		return "", err
	}

	return name, nil

}
