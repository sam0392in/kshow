// reference:
/*
	https://stackoverflow.com/questions/52029656/how-to-retrieve-kubernetes-metrics-via-client-go-and-golang
	https://github.com/kubernetes/metrics/tree/master/pkg/apis/metrics
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
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "NAMESPACE\t\tPOD\t\tCONTAINER\t\tCPU\t\tMemory")
	for _, m := range podMetricsList.Items {
		var (
			cpu float32
		)
		cpu = 0
		for _, c := range m.Containers {
			cpu = float32(c.Usage.Cpu().MilliValue())
			mem := c.Usage.Memory().Value() / 1048859
			data := m.Namespace + "\t\t" + m.Name + "\t\t" + c.Name + "\t\t" + strconv.Itoa(int(cpu)) + "m" + "\t\t" + strconv.Itoa(int(mem)) + "Mi"
			fmt.Fprintln(w, data)
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
