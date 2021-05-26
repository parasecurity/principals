package kubernetes

import (
	"log"

	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	config            *rest.Config
	clientset         *kubernetes.Clientset
	DeploymentsClient v1.DeploymentInterface
	kubeconfig        = "/home/.kube/config"
)

func loadDeployment() (result bool, err error) {

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Println("Error:", err)
		return false, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Println("Error:", err)
		return false, err
	}

	DeploymentsClient = clientset.AppsV1().Deployments("security")
	return true, err
}
