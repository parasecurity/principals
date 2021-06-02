package kubernetes

import (
	"context"
	"log"

	"api/pkg/yamls"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getDeployment(name string) appsv1.Deployment {
	var deployment appsv1.Deployment
	if name == "canary" {
		deployment = yamls.CreateCanaryDepl()
	} else if name == "detector" {
		deployment = yamls.CreateDetectorDepl()
	} else if name == "canary-link" {
		deployment = yamls.CreateCanaryLinkDepl()
	} else if name == "detector-link" {
		deployment = yamls.CreateDetectorLinkDepl()
	}
	return deployment
}

func createDeployment(command Command) {
	deployment := getDeployment(command.Name)
	log.Println("Creating deployment..")
	result, err := DeploymentsClient.Create(context.TODO(), &deployment, metav1.CreateOptions{})
	if err != nil {
		log.Println("Error on deployment:", err)
	}

	log.Println("Created deployment:", result.GetObjectMeta().GetName())
}

func Create(command Command) {
	if command.Name == "canary" ||
		command.Name == "detector" ||
		command.Name == "canary-link" ||
		command.Name == "detector-link" {
		_, err := loadDeployment()
		if err != nil {
			return
		}
		createDeployment(command)
	}

}
