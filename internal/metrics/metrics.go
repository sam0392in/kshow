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

// reference:
/*
	API:
	https://github.com/kubernetes/metrics/tree/master/pkg/apis/metrics

	Reference Implementation:
	https://stackoverflow.com/questions/52029656/how-to-retrieve-kubernetes-metrics-via-client-go-and-golang
*/

package metrics

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"text/tabwriter"

	"github.com/sam0392in/kshow/internal/pod"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

var (
	logger *zap.Logger
)

func init() {
	logger, _ = zap.NewProduction()

}

func client(namespace *string) (*metricsv.Clientset, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	}
	flag.Parse()
	_, err := os.Stat(*kubeconfig)
	config, _ := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	clientset, err := metricsv.NewForConfig(config)
	if err != nil {
		logger.Error(err.Error())
	}
	return clientset, err
}

// Get container resource usage
func PrintContainerMetrics(namespace string) {
	clientset, err := client(&namespace)
	podMetricsList, err := clientset.MetricsV1beta1().PodMetricses(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logger.Error(err.Error())
	}
	pods, err := pod.GetPods(&namespace)

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "NAMESPACE\t\tPOD\t\tCONTAINER\t\tCURRENT-CPU\t\tREQUESTED-CPU\t\tLIMIT-CPU\t\tCURRENT-MEMORY\t\tREQUESTED-MEMORY\t\tLIMIT-MEMORY")

	for _, m := range podMetricsList.Items {
		for _, p := range pods.Items {
			var (
				cpu                                            float32
				requestedCPU, requestedMem, LimitCPU, LimitMem string
			)
			if m.Name == p.Name {
				cpu = 0
				for _, c := range m.Containers {
					for _, c1 := range p.Spec.Containers {
						if c.Name == c1.Name {
							requestedCPU = c1.Resources.Requests.Cpu().String()
							requestedMem = c1.Resources.Requests.Memory().String()
							LimitCPU = c1.Resources.Limits.Cpu().String()
							LimitMem = c1.Resources.Limits.Memory().String()
							break
						}
					}
					cpu = float32(c.Usage.Cpu().MilliValue())
					mem := c.Usage.Memory().Value() / 1048859
					data := m.Namespace + "\t\t" + m.Name + "\t\t" + c.Name + "\t\t" + strconv.Itoa(int(cpu)) + "m" + "\t\t" + requestedCPU + "\t\t" + LimitCPU + "\t\t" + strconv.Itoa(int(mem)) + "Mi" + "\t\t" + requestedMem + "\t\t" + LimitMem
					fmt.Fprintln(w, data)
				}
			}
		}
	}
	w.Flush()
}

// Get Pod resource usage
func getPodMetrics(namespace *string) (*v1beta1.PodMetricsList, error) {
	clientset, err := client(namespace)
	podMetricsList, err := clientset.MetricsV1beta1().PodMetricses(*namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logger.Error(err.Error())
	}
	return podMetricsList, err
}

func PrintPodMetrics(namespace string) {
	podMetrics, err := getPodMetrics(&namespace)
	if err != nil {
		logger.Error(err.Error())
	}

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "NAMESPACE\t\tPOD\tCPU\t\tMEMORY")

	for _, m := range podMetrics.Items {
		var (
			cpu float32
			mem int64
		)
		cpu = 0
		mem = 0
		for _, c := range m.Containers {
			a := c.Usage.Cpu()
			cpu += float32(a.MilliValue())
			b := c.Usage.Memory().Value()
			mem += b
		}
		// Convert Ki to Mi matched with top command
		mem = mem / 1048859
		data := m.Namespace + "\t\t" + m.Name + "\t" + strconv.Itoa(int(cpu)) + "m" + "\t\t" + strconv.Itoa(int(mem)) + "Mi"
		fmt.Fprintln(w, data)
	}
	w.Flush()
}
