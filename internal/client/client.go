package client

import (
	"flag"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	logger *zap.Logger
)

func init() {
	logger, _ = zap.NewProduction()

}

func inCluster() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	return clientset, err
}

func outCluster(kubeconfig string) (*kubernetes.Clientset, error) {
	config, _ := clientcmd.BuildConfigFromFlags("", kubeconfig)
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	return clientset, err
}

func GetK8sClient() (*kubernetes.Clientset, error) {
	var (
		kubeconfig *string
		client     *kubernetes.Clientset
		err        error
	)
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	}
	flag.Parse()
	_, err = os.Stat(*kubeconfig)
	if err != nil {
		logger.Error(err.Error())
	}
	if os.IsNotExist(err) {
		client, err = inCluster()
	} else {
		client, err = outCluster(*kubeconfig)
	}

	return client, err
}
