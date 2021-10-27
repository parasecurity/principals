package yamls

import (
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateDetectorLinkDaem(args []string, registry *string) appsv1.DaemonSet {
	var image string = *registry + ":5000/antrea-tsi:v1.0.1"

	daemonSet := appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "detector-link",
			Namespace: "security",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":       "security",
					"component": "detector-link",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubectl.kubernetes.io/default-container": "detector-link",
						"k8s.v1.cni.cncf.io/networks": `[
							{ "name": "macvlan-conf",
							"ips": [ "10.1.1.203/24" ],
							"mac": "c2:b0:57:49:47:f1",
							"gateway": [ "10.1.1.1" ]
							}]`,
					},
					Labels: map[string]string{
						"app":       "security",
						"component": "detector-link",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "detector-link",
							Image: image,
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
							Command: []string{
								"/home/tsi/bin/detectorLink",
							},
							Args:            args,
							ImagePullPolicy: apiv1.PullAlways,
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "tsi-data",
									MountPath: "/home/data",
								},
								LogMount,
							},
						},
					},
					InitContainers: []apiv1.Container{
						mirrorContainter("detector", registry),
					},
					Volumes: []apiv1.Volume{
						RunAntreaVolume,
						{
							Name: "tsi-data",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/mnt/data/security_vol",
								},
							},
						},
						LogVolume,
					},
				},
			},
		},
	}

	return daemonSet
}
