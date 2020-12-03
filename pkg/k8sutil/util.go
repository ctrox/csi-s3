package k8sutil

import (
	log "github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func getRestConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}

	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func GetK8SClientSet(kubeconfig string) *kubernetes.Clientset {

	config, err := getRestConfig(kubeconfig)
	if err != nil {
		log.Fatalf("create kubenetes config fail: %s", err.Error())
		return nil
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("create kubenetes client set fail: %s", err.Error())
		return nil
	}

	return clientset
}

