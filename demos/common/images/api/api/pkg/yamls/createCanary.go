package yamls

import (
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateCanaryDepl() appsv1.Deployment {

	deployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "canary",
			Namespace: "security",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
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
							Image: "192.168.122.1:5000/tsi-tools:1.0.7",
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
							Args: []string{
								"-api=10.104.54.11:8001",
							},
						},
					},
				},
			},
		},
	}

	return deployment
}
