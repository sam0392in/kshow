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

package deployment

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"

	k8sclient "github.com/sam0392in/kshow/internal/client"
	"github.com/sam0392in/kshow/internal/node"
	"github.com/sam0392in/kshow/internal/pod"
	"go.uber.org/zap"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
)

var (
	logger *zap.Logger
)

func init() {
	logger, _ = zap.NewProduction()

}

func client(namespace *string) appsv1.DeploymentInterface {
	clientset, err := k8sclient.GetK8sClient()
	if err != nil {
		logger.Error(err.Error())
	}
	deployClient := clientset.AppsV1().Deployments(*namespace)
	return deployClient
}

/*
List Deployments,
Returns list.items of Deployments
*/
func GetDeployments(namespace *string) (*v1.DeploymentList, error) {
	deploymentsClient := client(namespace)
	list, err := deploymentsClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logger.Error(err.Error())
	}
	return list, err
}

// Extract Deployment name from pod name
func GetDeploymentFromPod(podName string) string {
	var deploymentName string
	pod := strings.Split(podName, "-")
	if len(pod) > 2 {
		podID := pod[len(pod)-1]
		replicasetID := pod[len(pod)-2]
		deployment := strings.Split(podName, "-"+replicasetID+"-"+podID)
		deploymentName = deployment[0]
	} else {
		podID := ""
		replicasetID := ""
		deployment := strings.Split(podName, "-"+replicasetID+"-"+podID)
		deploymentName = deployment[0]
	}

	return deploymentName
}

// Print deployments
func ListDeployments(namespace string) {
	deployList, err := GetDeployments(&namespace)
	if err != nil {
		logger.Error(err.Error())
	}
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "DEPLOYMENT\tNAMESPACE\tREPLICAS")
	for _, d := range deployList.Items {
		r := d.Spec.Replicas
		replicas := *r

		tr := d.Spec.Template.Spec.Tolerations
		var tolerations []string
		for _, t := range tr {
			tl := t.Key + "-" + string(t.Operator) + "-" + t.Value + "-" + string(t.Effect)
			tolerations = append(tolerations, tl)
		}

		data := d.Name + "\t" + d.Namespace + "\t" + strconv.FormatInt(int64(replicas), 10)
		fmt.Fprintln(w, data)

	}
	w.Flush()
}

// Get Pod extra details of Deployment
func getPodDistribution(namespace, deployment string) (int, int, int, int) {
	var podOnDemand, podSpot, podready, podtotal int
	podOnDemand = 0
	podSpot = 0
	podready = 0
	podtotal = 0
	pods, err := pod.GetPods(&namespace)
	if err != nil {
		logger.Error(err.Error())
	}
	// get all nodes in the cluster
	nodes, err := node.ListNodes()
	if err != nil {
		logger.Error(err.Error())
	}

	for _, pod := range pods.Items {
		podDeploymentName := GetDeploymentFromPod(pod.Name)

		re := regexp.MustCompile("^" + deployment + "$")
		exactMatch := re.MatchString(podDeploymentName)

		if exactMatch {
			podtotal += 1
			if pod.Status.Phase == "Running" {
				podready += 1
				for _, node := range nodes {
					podNode := pod.Spec.NodeName
					nodeName := node.ObjectMeta.Labels["kubernetes.io/hostname"]
					nodeTenancy := node.ObjectMeta.Labels["eks.amazonaws.com/capacityType"]
					if podNode == nodeName {
						if nodeTenancy == "ON_DEMAND" {
							podOnDemand += 1
						} else if nodeTenancy == "SPOT" {
							podSpot += 1
						}
						break
					}
				}
			}
		}

	}

	return podOnDemand, podSpot, podready, podtotal
}

// List deployments with Detailed
func ListDeploymentDetailed(namespace string) {
	deployList, err := GetDeployments(&namespace)
	if err != nil {
		logger.Error(err.Error())
	}
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "DEPLOYMENT\tNAMESPACE\t\tREADY\tDISTRIBUTION\t\tTOLERATIONS")
	for _, d := range deployList.Items {
		r := d.Spec.Replicas
		replicas := *r

		tr := d.Spec.Template.Spec.Tolerations
		var tolerations []string
		for _, t := range tr {
			tl := t.Key + "-" + string(t.Operator) + "-" + t.Value + "-" + string(t.Effect)
			tolerations = append(tolerations, tl)
		}

		// get distribution
		var distribution string
		// var percentageOndemand, percentageSpot float64
		podsOnDemand, podSpot, podReady, _ := getPodDistribution(namespace, d.Name)
		/*
			Disabled % distribution calculation.
			Currently enabled is distribution based on count of pods
		*/
		// totalpods := podsOnDemand + podSpot
		// if totalpods != 0 {
		// 	percentageOndemand = math.Round(float64(podsOnDemand * 100 / totalpods))
		// 	percentageSpot = math.Round(float64(podSpot * 100 / totalpods))
		// } else {
		// 	percentageOndemand = 0
		// 	percentageSpot = 0
		// }

		podReadyStatus := strconv.Itoa(podReady) + "/" + strconv.Itoa(int(replicas))

		// distribution = "OD:" + fmt.Sprintf("%.0f", percentageOndemand) + " SP:" + fmt.Sprintf("%.0f", percentageSpot)

		distribution = "OD:" + strconv.Itoa(podsOnDemand) + " SP:" + strconv.Itoa(podSpot)
		data := d.Name + "\t" + d.Namespace + "\t\t" + podReadyStatus + "\t" + distribution + "\t\t" + strings.Join(tolerations, "::")
		fmt.Fprintln(w, data)

	}
	w.Flush()
}
