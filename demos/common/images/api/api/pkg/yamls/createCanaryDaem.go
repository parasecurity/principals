package yamls

import (
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateCanaryDaem(args []string, registry *string) appsv1.DaemonSet {
	var image string = *registry + ":5000/antrea-tsi:v1.0.0"

	daemonSet := appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "canary",
			Namespace: "security",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":       "security",
					"component": "canary",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubectl.kubernetes.io/default-container": "canary",
						"k8s.v1.cni.cncf.io/networks":             "macvlan-host-local",
					},
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
							ImagePullPolicy: apiv1.PullAlways,
						},
					},
				},
			},
		},
	}

	return daemonSet
}
