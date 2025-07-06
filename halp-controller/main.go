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
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const mcpNamespace = "halpmcp"

func NewConfigMapInformer(clientset kubernetes.Interface, namespace string) cache.SharedIndexInformer {
	// Create a ListWatcher for ConfigMaps
	listWatcher := cache.NewListWatchFromClient(
		clientset.CoreV1().RESTClient(),
		"configmaps",
		namespace,
		fields.Everything(),
	)

	// Create the informer
	informer := cache.NewSharedIndexInformer(
		listWatcher,
		&corev1.ConfigMap{},
		time.Minute*10, // resync period
		cache.Indexers{},
	)

	// Add event handlers
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			cm, ok := obj.(*corev1.ConfigMap)
			if !ok {
				log.Fatalf("cannot convert to ConfigMap: %v", obj)
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
		},
	})

	return informer
}

func Run(ctx context.Context, informer cache.SharedIndexInformer) error {
	defer runtime.HandleCrash()

	log.Println("Starting ConfigMap controller")

	// Start the informer
	go informer.Run(ctx.Done())

	// Wait for the caches to sync
	if !cache.WaitForCacheSync(ctx.Done(), informer.HasSynced) {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	log.Println("ConfigMap controller synced and ready")

	<-ctx.Done()
	log.Println("Stopping ConfigMap controller")

	return nil
}

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

func main() {
	clientset, err := createKubernetesClient()
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	informer := NewConfigMapInformer(clientset, mcpNamespace)

	// Set up signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received termination signal, shutting down...")
		cancel()
	}()

	if err := Run(ctx, informer); err != nil {
		log.Fatalf("Error running controller: %v", err)
	}
}
