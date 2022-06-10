package yamls

import (
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateCanaryDepl(args []string, registry *string) appsv1.Deployment {
	var image string = *registry + antrea_image

	deployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "canary",
			Namespace: "security",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":       "security",
					"component": "canary",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":       "security",
						"component": "canary",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "canary",
							Image: image,
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
							Command: []string{
								"/home/tsi/bin/canary",
							},
							Args:            args,
							VolumeMounts: []apiv1.VolumeMount{
								LogMount,
							},
							ImagePullPolicy: apiv1.PullAlways,
						},
					},
					Volumes: []apiv1.Volume{
						LogVolume,
					},
				},
			},
		},
	}

	return deployment
}
