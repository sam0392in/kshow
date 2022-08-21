package pod

import (
	"context"
	k8sclient "kshow/internal/client"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	logger *zap.Logger
)

type ContainerDetails struct {
	Name, Currentcpu, Currentmemory string
}

type PodResourceStats struct {
	Name, Namespace string
	Containerstats  []ContainerDetails
}

func init() {
	logger, _ = zap.NewProduction()

}

func client(namespace *string) (*kubernetes.Clientset, error) {
	clientset, err := k8sclient.GetK8sClient()
	if err != nil {
		logger.Error(err.Error())
	}
	return clientset, err
}

// List Pods
func ListPods(namespace *string) (*v1.PodList, error) {
	clientset, err := client(namespace)
	if err != nil {
		logger.Error(err.Error())
	}
	podClient := clientset.CoreV1().Pods(*namespace)
	list, err := podClient.List(context.TODO(), metav1.ListOptions{})
	return list, err
}

// Get Requested CPU and Limit
// func GetRequestedResource(container, namespace *string) {
// 	pods, err := ListPods(namespace)
// 	if err != nil {
// 		logger.Error(err.Error())
// 	}

// }
