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

package node

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	k8sclient "github.com/sam0392in/kshow/internal/client"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	logger *zap.Logger
)

func init() {
	logger, _ = zap.NewProduction()

}

func client(namepsace string) *kubernetes.Clientset {
	clientset, err := k8sclient.GetK8sClient()
	if err != nil {
		logger.Error(err.Error())
	}
	return clientset
}

// returns the list of nodes in the cluster
func ListNodes() ([]v1.Node, error) {
	cl := client("")
	nodeclient := cl.CoreV1().Nodes()
	nodes, err := nodeclient.List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		return nil, err
	}
	return nodes.Items, nil
}

// Print List of Nodes
func GetNodeDetails() {
	nodes, err := ListNodes()
	if err != nil {
		logger.Error(err.Error())
	}
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "NODE\t\tSTATUS\t\tAGE\t\tVERSION")

	for _, n := range nodes {

		name := n.Name

		// node age
		var ageN string
		nodeCreated := n.CreationTimestamp
		age := time.Since(nodeCreated.Time).Round(time.Second)
		ageN = age.String()
		if age.Hours() > 8760 {
			ageInYears := int((age.Hours() + (age.Minutes() / 60)) / 8760)
			ageN = strconv.Itoa(ageInYears) + "y"
		} else if age.Hours() > 24 {
			ageInDays := int((age.Hours() + (age.Minutes() / 60)) / 24)
			ageN = strconv.Itoa(ageInDays) + "d"
		} else if age.Hours() > 1 {
			ageInHours := int(age.Hours() + (age.Minutes() / 60))
			ageN = strconv.Itoa(ageInHours) + "h"
		} else {
			ageInMin := int(age.Minutes() + (age.Seconds() / 60))
			ageN = strconv.Itoa(ageInMin) + "m"
		}

		// node status
		var nstatus string
		for _, s := range n.Status.Conditions {
			if s.Reason == "KubeletReady" {
				nstatus = string(s.Type)
			}
		}

		// Node k8s version
		nk8sVersion := n.Status.NodeInfo.KubeletVersion

		data := name + "\t\t" + nstatus + "\t\t" + ageN + "\t\t" + nk8sVersion
		fmt.Fprintln(w, data)
	}
	w.Flush()
}

// Print Detailed Node Info
func DetailedNodeInfo() {
	nodes, err := ListNodes()
	if err != nil {
		logger.Error(err.Error())
	}
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "NODE\t\tSTATUS\t\tAGE\t\tNODEGROUP\t\tTENANCY\t\tINSTANCE-TYPE\t\tARCH\t\tAWS-ZONE\t\tVERSION")

	for _, n := range nodes {

		name := n.Name

		// node age
		var ageN string
		nodeCreated := n.CreationTimestamp
		age := time.Since(nodeCreated.Time).Round(time.Second)
		ageN = age.String()
		if age.Hours() > 8760 {
			ageInYears := int((age.Hours() + (age.Minutes() / 60)) / 8760)
			ageN = strconv.Itoa(ageInYears) + "y"
		} else if age.Hours() > 24 {
			ageInDays := int((age.Hours() + (age.Minutes() / 60)) / 24)
			ageN = strconv.Itoa(ageInDays) + "d"
		} else if age.Hours() > 1 {
			ageInHours := int(age.Hours() + (age.Minutes() / 60))
			ageN = strconv.Itoa(ageInHours) + "h"
		} else {
			ageInMin := int(age.Minutes() + (age.Seconds() / 60))
			ageN = strconv.Itoa(ageInMin) + "m"
		}

		// node status
		var nstatus string
		for _, s := range n.Status.Conditions {
			if s.Reason == "KubeletReady" {
				nstatus = string(s.Type)
				if nstatus == "" {
					nstatus = "NotReady"
				}
			}
		}

		// nodegroup
		nodegroup := n.ObjectMeta.Labels["eks.amazonaws.com/nodegroup"]

		// Tenancy
		nodeTenancy := n.ObjectMeta.Labels["eks.amazonaws.com/capacityType"]

		// Instance Type
		nodeInstanceType := n.ObjectMeta.Labels["node.kubernetes.io/instance-type"]

		// Architecture
		arch := n.ObjectMeta.Labels["beta.kubernetes.io/arch"]

		// AWS Zone
		zone := n.ObjectMeta.Labels["topology.kubernetes.io/zone"]

		// Node k8s version
		nk8sVersion := n.Status.NodeInfo.KubeletVersion

		data := name + "\t\t" + nstatus + "\t\t" + ageN + "\t\t" + nodegroup + "\t\t" + nodeTenancy + "\t\t" + nodeInstanceType + "\t\t" + arch + "\t\t" + zone + "\t\t" + nk8sVersion
		fmt.Fprintln(w, data)
	}
	w.Flush()
}
