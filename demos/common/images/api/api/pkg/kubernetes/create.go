package kubernetes

import (
	"context"
	"log"

	"api/pkg/yamls"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getDeployment(name string, args []string) appsv1.Deployment {
	var deployment appsv1.Deployment
	if name == "canary" {
		deployment = yamls.CreateCanaryDepl(args)
	}

	return deployment
}

func getDaemonSet(name string, args []string) appsv1.DaemonSet {
	var daemonSet appsv1.DaemonSet
	if name == "detector-link" {
		daemonSet = yamls.CreateDetectorLinkDaem(args)
	} else if name == "canary-link" {
		daemonSet = yamls.CreateCanaryLinkDaem(args)
	} else if name == "detector" {
		daemonSet = yamls.CreateDetectorDaem(args)
	} else if name == "dga" {
		daemonSet = yamls.CreateDgaDaem(args)
	} else if name == "analyser" {
		daemonSet = yamls.CreateAnalyserDaem(args)
	} else if name == "snort" {
		daemonSet = yamls.CreateSnortDaem(args)
	} else if name == "honeypot" {
		daemonSet = yamls.CreateHoneypotDaem(args)
	}

	return daemonSet
}

func createDeployment(command Command) {
	deployment := getDeployment(command.Name, command.Arguments)
	log.Println("Creating deployment..")
	result, err := DeploymentsClient.Create(context.TODO(), &deployment, metav1.CreateOptions{})
	if err != nil {
		log.Println("Error on deployment:", err)
	}

	log.Println("Created deployment:", result.GetObjectMeta().GetName())
}

func createDaemonSet(command Command) {
	daemonset := getDaemonSet(command.Name, command.Arguments)
	log.Println("Creating DaemonSet..")
	result, err := DaemonSetClient.Create(context.TODO(), &daemonset, metav1.CreateOptions{})
	if err != nil {
		log.Println("Error on deamonset:", err)
	}

	log.Println("Created deamonset:", result.GetObjectMeta().GetName())
}

func Create(command Command) {
	if command.Name == "canary" {
		_, err := loadDeployment()
		if err != nil {
			return
		}

		createDeployment(command)
	} else if command.Name == "canary-link" ||
		command.Name == "detector-link" ||
		command.Name == "detector" ||
		command.Name == "dga" ||
		command.Name == "analyser" ||
		command.Name == "snort" ||
		command.Name == "honeypot" {
		_, err := loadDaemonSet()
		if err != nil {
			return
		}

		createDaemonSet(command)
	}

}
