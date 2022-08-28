package pod

import (
	"context"
	"fmt"
	k8sclient "kshow/internal/client"
	"kshow/internal/node"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

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

func GetPods(namespace *string) (*v1.PodList, error) {
	clientset, err := client(namespace)
	if err != nil {
		logger.Error(err.Error())
	}
	podClient := clientset.CoreV1().Pods(*namespace)
	list, err := podClient.List(context.TODO(), metav1.ListOptions{})
	return list, err
}

// List pods
func ListPods(namespace string) {
	pods, err := GetPods(&namespace)
	if err != nil {
		logger.Error(err.Error())
	}
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "POD\t\tREADY\t\tSTATUS\t\tRESTART\t\tAGE\t\tNAMESPACE")

	for _, pod := range pods.Items {

		// Calculate the age of the pod
		var ageS string
		podCreationTime := pod.GetCreationTimestamp()
		age := time.Since(podCreationTime.Time).Round(time.Second)
		ageS = age.String()
		if age.Hours() > 8760 {
			ageInYears := int((age.Hours() + (age.Minutes() / 60)) / 8760)
			ageS = strconv.Itoa(ageInYears) + "y"
		} else if age.Hours() > 24 {
			ageInDays := int((age.Hours() + (age.Minutes() / 60)) / 24)
			ageS = strconv.Itoa(ageInDays) + "d"
		} else if age.Hours() > 1 {
			ageInHours := int(age.Hours() + (age.Minutes() / 60))
			ageS = strconv.Itoa(ageInHours) + "h"
		} else {
			ageInMin := int(age.Minutes() + (age.Seconds() / 60))
			ageS = strconv.Itoa(ageInMin) + "m"
		}

		// Get the status of each of the pods
		podStatus := pod.Status

		var containerRestarts int32
		var containerReady int
		var totalContainers int

		// If a pod has multiple containers, get the status from all
		for container := range pod.Spec.Containers {
			containerRestarts += podStatus.ContainerStatuses[container].RestartCount
			if podStatus.ContainerStatuses[container].Ready {
				containerReady++
			}
			totalContainers++
		}

		// Get the values from the pod status
		name := pod.Name
		namespace := pod.Namespace
		ready := fmt.Sprintf("%v/%v", containerReady, totalContainers)
		status := fmt.Sprintf("%v", podStatus.Phase)
		restarts := fmt.Sprintf("%v", containerRestarts)

		data := name + "\t\t" + ready + "\t\t" + status + "\t\t" + restarts + "\t\t" + ageS + "\t\t" + namespace
		fmt.Fprintln(w, data)
	}
	w.Flush()
}

// List Pods with node tenancy (only for AWS EKS)
func ListPodswithNodeTenency(namespace string) {
	pods, err := GetPods(&namespace)
	if err != nil {
		logger.Error(err.Error())
	}
	// get all nodes in the cluster
	nodes, err := node.ListNodes()

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "POD\t\tSTATUS\t\tNAMESPACE\t\tNODE\t\tTENANCY")
	for _, pod := range pods.Items {
		for _, node := range nodes {
			podNode := pod.Spec.NodeName
			nodeName := node.ObjectMeta.Labels["kubernetes.io/hostname"]
			nodeTenancy := node.ObjectMeta.Labels["eks.amazonaws.com/capacityType"]
			if podNode == nodeName {
				data := pod.Name + "\t\t" + string(pod.Status.Phase) + "\t\t" + pod.Namespace + "\t\t" + podNode + "\t\t" + nodeTenancy
				fmt.Fprintln(w, data)
			}
		}
	}
	w.Flush()
}
