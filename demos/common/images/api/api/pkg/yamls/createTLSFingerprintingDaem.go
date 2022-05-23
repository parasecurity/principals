package yamls

import (
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateTLSFingerprintingDaem(args []string, registry *string) appsv1.DaemonSet {
	var dns_stitching_image string = ":5000/tls-fingerprint:v0.1.0"
	var image string = *registry + dns_stitching_image

	daemonSet := appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tls-fingerprint",
			Namespace: "security",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":       "security",
					"component": "tls-fingerprint",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubectl.kubernetes.io/default-container": "tls-fingerprint",
						"k8s.v1.cni.cncf.io/networks": `[
							{ "name": "macvlan-conf",
							"ips": [ "10.1.1.204/24" ],
							"mac": "c2:b0:57:49:47:f1",
							"gateway": [ "10.1.1.1" ]
							}]`,
					},
					Labels: map[string]string{
						"app":       "security",
						"component": "tls-fingerprint",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "tls-fingerprint",
							Image: image,
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 8020,
								},
							},
							Command: []string{
								"/home/tsi/bin/tls-fingerprint",
							},
							Args:            args,
							VolumeMounts: []apiv1.VolumeMount{
								LogMount,
							},
							ImagePullPolicy: apiv1.PullAlways,
						},
					},
					InitContainers: []apiv1.Container{
						mirrorContainter("tls-fing", registry),
					},
					Volumes: []apiv1.Volume{
						RunAntreaVolume,
						LogVolume,
					},
				},
			},
		},
	}

	return daemonSet
}
