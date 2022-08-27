package main

import (
	"kshow/internal/deployment"
	"kshow/internal/metrics"
	"kshow/internal/pod"
	"os"

	"go.uber.org/zap"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	logger *zap.Logger

	app = kingpin.New("kshow", "A command-line tool for kubernetes.")

	get            = app.Command("get", "k8s object type list")
	k8sObject      = get.Arg("k8s object", "Specify k8s object").Required().String()
	namespace      = get.Flag("namespace", "Specify namespace. default is all namespace").Default("").String()
	showToleration = get.Flag("show-tolerations", "show toleration of deloyment").Bool()
	// Pod Specific Arguments
	showNodeTenancy = get.Flag("show-tenancy", "Show tenancy of node. Specific to AWS").Bool()

	resourceStats   = app.Command("resource-stats", "Show resource statistics")
	statsNamespace  = resourceStats.Flag("namespace", "Specify namespace. default is all namespace").Default("").String()
	statsContainers = resourceStats.Flag("containers", "").Bool()
)

func init() {
	logger, _ = zap.NewProduction()

}

func getDeployments() {
	if *showToleration {
		deployment.ListDeploymentwithTolerations(*namespace)
	} else {
		deployment.ListDeployments(*namespace)
	}
}

func getPods() {
	if *showNodeTenancy {
		pod.ListPodswithNodeTenency(*namespace)
	} else {
		pod.ListPods(*namespace)
	}
}

func getMetrics() {
	if *statsContainers {
		metrics.PrintContainerMetrics(*statsNamespace)
	} else {
		metrics.PrintPodMetrics(*statsNamespace)
	}
}

func getObject() {
	switch *k8sObject {
	case "deployment", "deployments", "deploy":
		getDeployments()
	case "pods", "pod":
		getPods()
	}
}

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case get.FullCommand():
		getObject()
	case resourceStats.FullCommand():
		getMetrics()
	}
}
