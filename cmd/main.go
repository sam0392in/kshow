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

package main

import (
	"os"

	"github.com/sam0392in/kshow/internal/deployment"
	"github.com/sam0392in/kshow/internal/metrics"
	"github.com/sam0392in/kshow/internal/node"
	"github.com/sam0392in/kshow/internal/pod"

	"go.uber.org/zap"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	logger *zap.Logger

	app = kingpin.New("kshow", "A command-line tool for kubernetes.")

	get            = app.Command("get", "get details of kubernetes objects")
	k8sObject      = get.Arg("k8s object", "allowed objects: deployment, pods").Required().String()
	namespace      = get.Flag("namespace", "Specify namespace. default is all namespace").Short('n').Default("").String()
	showToleration = get.Flag("show-tolerations", "show tolerations of a deloyment").Bool()
	// Pod and Node Specific Argument
	detailed = get.Flag("detailed", "Show extra details").Bool()

	resourceStats  = app.Command("resource-stats", "Show current resource statistics")
	statsNamespace = resourceStats.Flag("namespace", "Specify namespace. default is all namespace").Short('n').Default("").String()
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

func getTest() {
	metrics.GetTotalClusterResources()
}

func getObject() {
	switch *k8sObject {
	case "deployment", "deployments", "deploy":
		getDeployments()
	case "pods", "pod", "po":
		getPods()
	case "node", "nodes", "no":
		getNodes()
	case "test":
		getTest()
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
