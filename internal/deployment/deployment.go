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
	k8sclient "kshow/internal/client"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

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
func getDeployments(namespace *string) (*v1.DeploymentList, error) {
	deploymentsClient := client(namespace)
	list, err := deploymentsClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logger.Error(err.Error())
	}
	return list, err
}

// Print deployments
func ListDeployments(namespace string) {
	deployList, err := getDeployments(&namespace)
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

// List deployments with tolerations
func ListDeploymentwithTolerations(namespace string) {
	deployList, err := getDeployments(&namespace)
	if err != nil {
		logger.Error(err.Error())
	}
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "DEPLOYMENT\tNAMESPACE\tREPLICAS\tTOLERATIONS")
	for _, d := range deployList.Items {
		r := d.Spec.Replicas
		replicas := *r

		tr := d.Spec.Template.Spec.Tolerations
		var tolerations []string
		for _, t := range tr {
			tl := t.Key + "-" + string(t.Operator) + "-" + t.Value + "-" + string(t.Effect)
			tolerations = append(tolerations, tl)
		}

		data := d.Name + "\t" + d.Namespace + "\t" + strconv.FormatInt(int64(replicas), 10) + "\t" + strings.Join(tolerations, "::")
		fmt.Fprintln(w, data)

	}
	w.Flush()
}
