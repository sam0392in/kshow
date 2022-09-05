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
	logger      *zap.Logger
	lineBreaker string
)

func init() {
	logger, _ = zap.NewProduction()
	lineBreaker = "--------------------------------------------------------------------------------------------------------------------------------------"

}

func client(namepsace string) *kubernetes.Clientset {
	clientset, err := k8sclient.GetK8sClient()
	if err != nil {
		fmt.Println(err.Error())
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

// Get Nodegroups
func getNodeGroups(nodes []v1.Node) []string {
	keys := make(map[string]bool)
	uniqueNG := []string{}
	for _, n := range nodes {
		ng := n.ObjectMeta.Labels["eks.amazonaws.com/nodegroup"]
		if _, value := keys[ng]; !value {
			keys[ng] = true
			uniqueNG = append(uniqueNG, ng)
		}
	}
	return uniqueNG
}

// Get Count of Node per Nodegroup
func getNodeCountPerNG(nodes []v1.Node) map[string]int {
	ngCount := make(map[string]int)
	nglist := getNodeGroups(nodes)
	for _, ng := range nglist {
		count := 0
		for _, n := range nodes {
			specifiedNG := n.ObjectMeta.Labels["eks.amazonaws.com/nodegroup"]
			if ng == specifiedNG {
				count += 1
			}
		}
		ngCount[ng] = count
	}
	return ngCount
}

// Get ClusterVersion
func getClusterVersion(nodes []v1.Node) []string {
	keys := make(map[string]bool)
	versions := []string{}
	for _, n := range nodes {
		k8sVersion := n.Status.NodeInfo.KubeletVersion
		if _, value := keys[k8sVersion]; !value {
			keys[k8sVersion] = true
			versions = append(versions, k8sVersion)
		}
	}
	return versions
}

// Nod Output Header
func NodeHeader(nodes []v1.Node) {
	//get clusterversion
	k8sVersion := getClusterVersion(nodes)
	// Get NodeGroup Node Count
	ngDetails := getNodeCountPerNG(nodes)

	// Print Header
	fmt.Println(lineBreaker)
	fmt.Println("K8S-VERSION\t\t\tNODE-GROUP: NODECOUNT")
	i := 0
	j := len(k8sVersion)
	for ng := range ngDetails {
		if i < j {
			fmt.Println(k8sVersion[i] + "\t\t" + ng + ":  " + strconv.Itoa(ngDetails[ng]))
		} else {
			fmt.Println("\t\t\t\t" + ng + ":  " + strconv.Itoa(ngDetails[ng]))
		}
		i++
	}
	fmt.Println(lineBreaker)
}

// Print List of Nodes
func GetNodeDetails() {
	nodes, _ := ListNodes()

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
	nodes, _ := ListNodes()

	NodeHeader(nodes)

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "\nNODE\t\tSTATUS\t\tAGE\t\tNODEGROUP\t\tTENANCY\t\tINSTANCE-TYPE\t\tARCH\t\tAWS-ZONE")

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

		data := name + "\t\t" + nstatus + "\t\t" + ageN + "\t\t" + nodegroup + "\t\t" + nodeTenancy + "\t\t" + nodeInstanceType + "\t\t" + arch + "\t\t" + zone + "\t\t"
		fmt.Fprintln(w, data)
	}
	w.Flush()
}
