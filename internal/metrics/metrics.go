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
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"text/tabwriter"

	"github.com/sam0392in/kshow/internal/node"
	"github.com/sam0392in/kshow/internal/pod"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

var (
	logger      *zap.Logger
	lineBreaker string
)

func init() {
	logger, _ = zap.NewProduction()
	lineBreaker = "--------------------------------------------------------------------------------------------------------------------------------------------------------"

}

func client(namespace *string) (*metricsv.Clientset, error) {
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}
	_, err := os.Stat(kubeconfig)
	config, _ := clientcmd.BuildConfigFromFlags("", kubeconfig)
	clientset, err := metricsv.NewForConfig(config)
	if err != nil {
		logger.Error(err.Error())
	}
	return clientset, err
}

// Get total CPU and MEM of the cluster
func GetTotalClusterResources() (float64, float64) {
	var cpu, mem float64
	nodes, err := node.ListNodes()
	if err != nil {
		logger.Error(err.Error())
	}

	for _, n := range nodes {
		cpu += n.Status.Allocatable.Cpu().AsApproximateFloat64()
		mem += (n.Status.Allocatable.Memory().AsApproximateFloat64()) / 1048859000
	}
	return cpu, mem
}

// Get total CPU and MEM of the namespace
func GetTotalNamespaceResources(namespace string) (float64, float64) {
	var (
		nsCPU, nsMEM float64
	)
	nsCPU = 0
	nsMEM = 0
	clientset, err := client(&namespace)
	podMetricsList, err := clientset.MetricsV1beta1().PodMetricses(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logger.Error(err.Error())
	}
	pods, err := pod.GetPods(&namespace)
	if err != nil {
		logger.Error(err.Error())
	}

	// for _, p := range pods.Items {
	// 	for _, c := range p.Spec.Containers {

	// 		reqcpu := (c.Resources.Requests.Cpu().AsApproximateFloat64())
	// 		reqmem := c.Resources.Requests.Memory().AsApproximateFloat64() / 1048859000
	// 		curentcpu := c.
	// 		cpu := float32(c.Usage.Cpu().MilliValue())
	// 		mem := c.Usage.Memory().Value() / 1048859
	// 	}
	// }
	for _, m := range podMetricsList.Items {
		for _, p := range pods.Items {
			var (
				cpu, mem, requestedCPU, requestedMem float64
			)
			if m.Name == p.Name {
				for _, c := range m.Containers {
					for _, c1 := range p.Spec.Containers {
						if c.Name == c1.Name {
							requestedCPU = c1.Resources.Requests.Cpu().AsApproximateFloat64()
							requestedMem = c1.Resources.Requests.Memory().AsApproximateFloat64() / 1048859000
							break
						}
					}
					cpu = float64(c.Usage.Cpu().AsApproximateFloat64())
					mem = float64(c.Usage.Memory().AsApproximateFloat64() / 1048859000)

				}
				// CPU and Memory will be taken in account which ever is higher of Requested VS Current
				if cpu > requestedCPU {
					nsCPU += cpu
				} else {
					nsCPU += requestedCPU
				}

				if mem > requestedMem {
					nsMEM += mem
				} else {
					nsMEM += requestedMem
				}
			}
		}
	}
	return nsCPU, nsMEM
}

// Get container resource usage
func PrintContainerMetrics(namespace string) {
	clientset, err := client(&namespace)
	podMetricsList, err := clientset.MetricsV1beta1().PodMetricses(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logger.Error(err.Error())
	}
	pods, err := pod.GetPods(&namespace)

	// Get Total Cluster stats
	totalCPU, totalMem := GetTotalClusterResources()

	// Get Total NS Stats
	nsCPU, nsMem := GetTotalNamespaceResources(namespace)

	// Get % Stats
	perCPU := (nsCPU / totalCPU) * 100
	perMEM := (nsMem / totalMem) * 100

	// Print Header
	fmt.Println(lineBreaker)
	fmt.Println("Cluster Stats: \t\tTotal CPU: " + strconv.Itoa(int(totalCPU)) + " Cores\t\tTotal Memory: " + strconv.Itoa(int(totalMem)) + " GB")
	fmt.Println("Namespace Stats: \tConsumed CPU: " + strconv.Itoa(int(nsCPU)) + " Cores\t\tConsumed Memory: " + strconv.Itoa(int(nsMem)) + " GB")
	fmt.Println("% Stats: \t\tCPU: " + fmt.Sprintf("%.2f", perCPU) + " %\t\t\tMemory: " + fmt.Sprintf("%.2f", perMEM) + " %")
	fmt.Println(lineBreaker)

	// Container stats
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "NAMESPACE\t\tPOD\t\tCONTAINER\t\tCURRENT-CPU\t\tREQ-CPU\t\tLIMIT-CPU\t\tCURRENT-MEM\t\tREQ-MEM\t\tLIMIT-MEM")

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
