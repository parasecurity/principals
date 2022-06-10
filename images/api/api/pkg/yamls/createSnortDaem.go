package yamls

import (
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateSnortDaem(args []string, registry *string) appsv1.DaemonSet {
	var HostPathDirectoryOrCreate apiv1.HostPathType = "DirectoryOrCreate"
	var image string = *registry + ":5000/tsi-snort:v1.0.0"

	daemonSet := appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "snort",
			Namespace: "security",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":       "security",
					"component": "snort",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubectl.kubernetes.io/default-container": "snort",
						"k8s.v1.cni.cncf.io/networks": `[
							{ "name": "macvlan-conf",
								"ips": [ "10.1.1.102/24" ],
								"mac": "c2:b0:57:49:47:f1",
								"gateway": [ "10.1.1.1" ]
						}]`,
					},
					Labels: map[string]string{
						"app":       "security",
						"component": "snort",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "snort",
							Image: image,
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
							Command: []string{
								"snort",
							},
							Args:            args,
							ImagePullPolicy: apiv1.PullAlways,
						},
					},
					InitContainers: []apiv1.Container{
						mirrorContainter("snort", registry),
					},
					Volumes: []apiv1.Volume{
						{
							Name: "host-var-run-antrea",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/var/run/antrea",
									Type: &HostPathDirectoryOrCreate,
								},
							},
						},
					},
				},
			},
		},
	}

	return daemonSet
}
