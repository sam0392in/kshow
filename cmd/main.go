package main

import (
	"kshow/internal/deployment"
	"kshow/internal/metrics"
	"kshow/internal/node"
	"kshow/internal/pod"
	"os"

	"go.uber.org/zap"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	logger *zap.Logger

	app = kingpin.New("kshow", "A command-line tool for kubernetes.")

	get            = app.Command("get", "get details of kubernetes objects")
	k8sObject      = get.Arg("k8s object", "allowed objects: deployment, pods").Required().String()
	namespace      = get.Flag("namespace", "Specify namespace. default is all namespace").Default("").String()
	showToleration = get.Flag("show-tolerations", "show tolerations of a deloyment").Bool()
	// Pod and Node Specific Argument
	detailed = get.Flag("detailed", "Show extra details").Bool()

	resourceStats  = app.Command("resource-stats", "Show current resource statistics")
	statsNamespace = resourceStats.Flag("namespace", "Specify namespace. default is all namespace").Default("").String()
	statsDetailed  = resourceStats.Flag("detailed", "show detailed resource statistics").Bool()
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
	if *detailed {
		pod.ListPodswithNodeTenency(*namespace)
	} else {
		pod.ListPods(*namespace)
	}
}

func getNodes() {
	if *detailed {
		node.DetailedNodeInfo()
	} else {
		node.GetNodeDetails()
	}
}

func getMetrics() {
	if *statsDetailed {
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
	case "node", "nodes":
		getNodes()
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
