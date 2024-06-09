package k8sclient

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func InitClients() (*kubernetes.Clientset, *dynamic.DynamicClient, error) {
	// must run in a kube cluster otherwise will error
	kubeconfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, nil, err
	}

	clientSet, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return nil, nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(kubeconfig)
	if err != nil {
		return nil, nil, err
	}

	return clientSet, dynamicClient, nil
}
