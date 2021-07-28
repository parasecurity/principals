package kubernetes

import (
	"context"
	"log"

	"api/pkg/yamls"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Run(command Command, registry *string) {
	_, err := loadDaemonSet()
	if err != nil {
		return
	}

	yaml := yamls.CreateRunYaml(command.Target, command.Arguments, registry)
	log.Println("Creating Run Yaml..")
	result, err := DaemonSetClient.Create(context.TODO(), &yaml, metav1.CreateOptions{})
	if err != nil {
		log.Println("Error on deamonset:", err)
	}

	log.Println("Created deamonset:", result.GetObjectMeta().GetName())
}
