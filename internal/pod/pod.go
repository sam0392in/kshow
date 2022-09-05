/*
Copyright 2022 Samarth Kanungo.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pod

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	k8sclient "github.com/sam0392in/kshow/internal/client"
	"github.com/sam0392in/kshow/internal/node"

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

// Get the Age of pod
func getPodAge(podCreationTime metav1.Time) string {
	var ageS string
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
	return ageS
}

// Get Pod status
func getPodStatus(podStatus v1.PodStatus) string {
	var status string
	if podStatus.Phase != "Running" && podStatus.Phase != "Succeeded" && len(podStatus.ContainerStatuses) != 0 {
		for _, cs := range podStatus.ContainerStatuses {
			if !(cs.Ready) && (cs.State.Waiting != nil) {
				status = cs.State.Waiting.Reason
				break
			} else {
				status = string(podStatus.Phase)
			}
		}
	} else {
		status = string(podStatus.Phase)
	}
	return status
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
		ageS = getPodAge(podCreationTime)

		// Get the status of each of the pods
		podStatus := pod.Status
		var status string
		status = getPodStatus(podStatus)

		var containerRestarts int32
		var containerReady int
		var totalContainers int

		// If a pod has multiple containers, get the status from all
		for _, s := range pod.Status.ContainerStatuses {
			containerRestarts += s.RestartCount
			if s.Ready {
				containerReady++
			}
			totalContainers++
		}

		// Get the values from the pod status
		name := pod.Name
		namespace := pod.Namespace
		ready := fmt.Sprintf("%v/%v", containerReady, totalContainers)
		// status := fmt.Sprintf("%v", podStatus.)
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
	fmt.Fprintln(w, "POD\t\tAGE\t\tSTATUS\t\tNAMESPACE\t\tNODE\t\tTENANCY")
	for _, pod := range pods.Items {
		for _, node := range nodes {
			podNode := pod.Spec.NodeName
			nodeName := node.ObjectMeta.Labels["kubernetes.io/hostname"]
			nodeTenancy := node.ObjectMeta.Labels["eks.amazonaws.com/capacityType"]
			if podNode == nodeName {

				// Calculate the age of the pod
				var ageS string
				podCreationTime := pod.GetCreationTimestamp()
				ageS = getPodAge(podCreationTime)

				// Get the status of each of the pods
				podStatus := pod.Status
				var status string
				status = getPodStatus(podStatus)

				data := pod.Name + "\t\t" + ageS + "\t\t" + status + "\t\t" + pod.Namespace + "\t\t" + podNode + "\t\t" + nodeTenancy
				fmt.Fprintln(w, data)
			}
		}
	}
	w.Flush()
}
