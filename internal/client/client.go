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

package client

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	logger *zap.Logger
)

func init() {
	logger, _ = zap.NewProduction()

}

func inCluster() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	return clientset, err
}

func outCluster(kubeconfig string) (*kubernetes.Clientset, error) {
	config, _ := clientcmd.BuildConfigFromFlags("", kubeconfig)
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	return clientset, err
}

func GetK8sClient() (*kubernetes.Clientset, error) {
	var (
		kubeconfig string
		client     *kubernetes.Clientset
		err        error
	)
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}
	_, err = os.Stat(kubeconfig)
	if err != nil {
		logger.Error(err.Error())
	}
	if os.IsNotExist(err) {
		client, err = inCluster()
	} else {
		client, err = outCluster(kubeconfig)
	}
	return client, err
}
