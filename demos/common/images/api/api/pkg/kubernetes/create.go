package kubernetes

import (
	"context"
	log "logging"
	"strings"

	"api/pkg/yamls"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getCommand(args []string) string {
	argsLength := len(args)
	for counter := 0; counter < argsLength; counter++ {
		value := args[counter]
		isCommand := strings.Contains(value, "-c=") || strings.Contains(value, "-command=")
		if isCommand {
			valueArray := strings.Split(value, "=")
			return valueArray[1]
		}
	}

	return "nil"
}

func getDeployment(name string, args []string, registry *string) appsv1.Deployment {
	var deployment appsv1.Deployment
	if name == "canary" {
		deployment = yamls.CreateCanaryDepl(args, registry)
	}

	return deployment
}

func getDaemonSet(name string, args []string, registry *string) appsv1.DaemonSet {
	var daemonSet appsv1.DaemonSet
	if name == "canary" {
		daemonSet = yamls.CreateCanaryDaem(args, registry)
	} else if name == "detector-link" {
		daemonSet = yamls.CreateDetectorLinkDaem(args, registry)
	} else if name == "canary-link" {
		daemonSet = yamls.CreateCanaryLinkDaem(args, registry)
	} else if name == "detector" {
		daemonSet = yamls.CreateDetectorDaem(args, registry)
	} else if name == "dga" {
		command := getCommand(args)
		if command == "forward" {
			daemonSet = yamls.CreateDgaForwardDaem(args, registry)
		} else {
			daemonSet = yamls.CreateDgaDaem(args, registry)
		}
	} else if name == "analyser" {
		daemonSet = yamls.CreateAnalyserDaem(args, registry)
	} else if name == "snort" {
		daemonSet = yamls.CreateSnortDaem(args, registry)
	} else if name == "honeypot" {
		daemonSet = yamls.CreateHoneypotDaem(args, registry)
	}

	return daemonSet
}

func createDeployment(command Command, registry *string) {
	deployment := getDeployment(command.Target, command.Arguments, registry)
	log.Println("Creating deployment..")
	result, err := DeploymentsClient.Create(context.TODO(), &deployment, metav1.CreateOptions{})
	if err != nil {
		log.Println("Error on deployment:", err)
	}

	log.Println("Created deployment:", result.GetObjectMeta().GetName())
}

func createDaemonSet(command Command, registry *string) {
	daemonset := getDaemonSet(command.Target, command.Arguments, registry)
	log.Println("Creating DaemonSet..")
	result, err := DaemonSetClient.Create(context.TODO(), &daemonset, metav1.CreateOptions{})
	if err != nil {
		log.Println("Error on deamonset:", err)
	}

	log.Println("Created deamonset:", result.GetObjectMeta().GetName())
}

func Create(command Command, registry *string) {
	if command.Target == "canary" ||
		command.Target == "canary-link" ||
		command.Target == "detector-link" ||
		command.Target == "detector" ||
		command.Target == "dga" ||
		command.Target == "analyser" ||
		command.Target == "snort" ||
		command.Target == "honeypot" {
		_, err := loadDaemonSet()
		if err != nil {
			return
		}

		createDaemonSet(command, registry)
	}

}
