package main

import (
	"context"
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const configMapName = "halp-message-config-map"

// Create a config map with the desired message in its data field.
// subsequently, delete the config map too for subsequent re-use.
func createAndDeleteConfigMap(ctx context.Context, message, ns string) (*corev1.ConfigMap, error) {
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %v", err)
	}

	// Check if ns exists; if not create it.
	_, err = clientset.CoreV1().Namespaces().Get(ctx, ns, metav1.GetOptions{})

	if k8serrors.IsNotFound(err) {
		toCreate := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: ns},
		}
		_, err = clientset.CoreV1().Namespaces().Create(ctx, toCreate, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: ns,
		},
		Data: map[string]string{
			"halpMessage": message,
		},
	}

	result, err := clientset.CoreV1().ConfigMaps(ns).Create(
		ctx,
		configMap,
		metav1.CreateOptions{},
	)

	if err != nil {
		return nil, err
	}
	log.Printf("Created ConfigMap: %s in namespace: %s", configMapName, ns)

	err = clientset.CoreV1().ConfigMaps(ns).Delete(ctx, configMapName, metav1.DeleteOptions{})
	if err != nil {
		return nil, err
	}
	log.Printf("Deleted ConfigMap: %s in namespace: %s", configMapName, ns)

	return result, nil
}
