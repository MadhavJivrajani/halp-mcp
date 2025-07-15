package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const mcpNamespace = "halpmcp"

// createKubernetesClient creates a Kubernetes client
func createKubernetesClient() (kubernetes.Interface, error) {
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %v", err)
	}

	return clientset, nil
}

func handleEvent(event watch.Event) {
	switch event.Type {
	case watch.Added:
		cm, ok := event.Object.(*corev1.ConfigMap)
		if !ok {
			panic(fmt.Sprintf("received object not a ConfigMap: %#v", event.Object))
		}
		// Ideally we could create a CRD maybe and have fun with that.
		// But since we are doing configmaps, creating a new NS adds in
		// the root certs CM to it leading to events that are valid,
		// but not of use to us. Skip those.
		if !strings.Contains(cm.Name, "halp") {
			log.Printf("irrelevant configmap %s, skipping.\n", cm.Name)
			return
		}
		msg := cm.Data["halpMessage"]
		log.Println("ConfigMap data:", cm.Data["halpMessage"])
		if err := exec.Command("halp", "-m", msg).Run(); err != nil {
			log.Fatalf("error trying to execute the halp command: %v", err)
		}
	}
}

func main() {
	clientset, err := createKubernetesClient()
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received termination signal, shutting down...")
		cancel()
	}()

	w, _ := clientset.CoreV1().ConfigMaps(mcpNamespace).Watch(ctx, v1.ListOptions{Watch: true, ResourceVersion: ""})
	watchChan := w.ResultChan()
	for {
		select {
		case event := <-watchChan:
			handleEvent(event)
		case <-ctx.Done():
			return
		}
	}
}
