package yamls

import (
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateDetectorDepl(args []string, registry *string) appsv1.Deployment {
	var HostPathDirectoryOrCreate apiv1.HostPathType = "DirectoryOrCreate"
	var image string = *registry + ":5000/antrea-tsi:v1.0.0"

	deployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "detector",
			Namespace: "security",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":       "security",
					"component": "detector",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubectl.kubernetes.io/default-container": "detector",
						"k8s.v1.cni.cncf.io/networks":             "macvlan-host-local",
					},
					Labels: map[string]string{
						"app":       "security",
						"component": "detector",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "detector",
							Image: image,
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
							Command: []string{
								"/home/tsi/bin/detector",
							},
							Args:            args,
							ImagePullPolicy: apiv1.PullAlways,
						},
					},
					InitContainers: []apiv1.Container{
						{
							Name:  "init-mirror",
							Image: image,
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
