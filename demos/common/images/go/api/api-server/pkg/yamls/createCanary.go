package yamls

import (
	"api/pkg/utils"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateCanaryDepl() appsv1.Deployment {

	deployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "canary",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: utils.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "canary",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "canary",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "canary",
							Image: "192.168.122.1:5000/tsi-tools:1.0.0",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
							Command: []string{
								"./httpCanary",
							},
						},
					},
				},
			},
		},
	}

	return deployment
}
