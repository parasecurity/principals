package kubernetes

import (
	"context"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deleteDeployment(name string) {
	log.Println("Deleting deployment...")
	deletePolicy := metav1.DeletePropagationForeground
	if err := DeploymentsClient.Delete(context.TODO(), name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		log.Println("Error on deletion:", err)
		return
	}
	log.Println("Deleted Deployment", name)
}

func deleteDaemonset(name string) {
	log.Println("Deleting DaemonSet...")
	deletePolicy := metav1.DeletePropagationForeground
	if err := DaemonSetClient.Delete(context.TODO(), name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		log.Println("Error on deletion:", err)
		return
	}
	log.Println("Deleted DeamonSet", name)
}

func Delete(command Command) {
	if command.Name == "canary" {
		_, err := loadDeployment()
		if err != nil {
			return
		}

		deleteDeployment(command.Name)
	} else if command.Name == "canary-link" ||
		command.Name == "detector-link" ||
		command.Name == "detector" ||
		command.Name == "dga" ||
		command.Name == "analyser" ||
		command.Name == "snort" {
		_, err := loadDaemonSet()
		if err != nil {
			return
		}

		deleteDaemonset(command.Name)
	}

}
