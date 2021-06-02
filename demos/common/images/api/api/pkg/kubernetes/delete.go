package kubernetes

import (
	"context"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deleteDeployment(name string) {
	log.Println("Deleting primitive...")
	deletePolicy := metav1.DeletePropagationForeground
	if err := DeploymentsClient.Delete(context.TODO(), name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		log.Println("Error on deletion:", err)
	}
	log.Println("Deleted primitive.")
}

func Delete(command Command) {
	if command.Name == "canary" ||
		command.Name == "detector" ||
		command.Name == "canary-link" ||
		command.Name == "detector-link" {
		_, err := loadDeployment()
		if err != nil {
			return
		}
		deleteDeployment(command.Name)
	}

}
