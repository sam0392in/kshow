package node

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

func init() {
	logger, _ = zap.NewProduction()

}

func client(namepsace string) *kubernetes.Clientset {
	clientset, err := k8sclient.GetK8sClient()
	if err != nil {
		logger.Error(err.Error())
	}
	return clientset
}

// returns the list of nodes in the cluster
func ListNodes() ([]v1.Node, error) {
	cl := client("")
	nodeclient := cl.CoreV1().Nodes()
	nodes, err := nodeclient.List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		return nil, err
	}
	return nodes.Items, nil
}
