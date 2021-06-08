package yamls

import (
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateDetectorDepl(args []string) appsv1.Deployment {
	var HostPathDirectoryOrCreate apiv1.HostPathType = "DirectoryOrCreate"

	deployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "detector",
			Namespace: "security",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "detector",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubectl.kubernetes.io/default-container": "detector",
						"k8s.v1.cni.cncf.io/networks": `[
							{ "name": "macvlan-conf",
								"ips": [ "10.1.1.102/24" ],
								"mac": "c2:b0:57:49:47:f1",
								"gateway": [ "10.1.1.1" ]
						}]`,
					},
					Labels: map[string]string{
						"app": "detector",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "detector",
							Image: "130.207.224.36:5000/tsi-tools:1.0.22",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
							Command: []string{
								"./home/httpDetector",
							},
							Args:            args,
							ImagePullPolicy: apiv1.PullAlways,
						},
					},
					InitContainers: []apiv1.Container{
						{
							Name:  "init-mirror",
							Image: "130.207.224.36:5000/antrea-tsi:v1.0.1",
							Env: []apiv1.EnvVar{
								{
									Name:  "NAME",
									Value: "detector",
								},
							},
							Command: []string{
								"sh",
								"-c",
								"/home/tsi/scripts/mirror-port.sh",
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "host-var-run-antrea",
									MountPath: "/var/run/openvswitch",
									SubPath:   "openvswitch",
								},
							},
						},
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

	return deployment
}
