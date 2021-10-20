package yamls

import (
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateRunYaml(code string, args []string, registry *string) appsv1.DaemonSet {
	// Parsing the correct registry IP address
	var image string = *registry + ":5000/antrea-tsi:v1.0.1"
	var HostPathDirectoryOrCreate apiv1.HostPathType = "DirectoryOrCreate"
	var sudo int64 = int64(0)
	var trueVar bool = true

	/* TODO: Maybe do any code modification here or any parsing
	*  TODO: Maybe pass an option to check if the admin wants to
	*  create a daemonset or a deployment
	 */
	daemonSet := appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "executor",
			Namespace: "security",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":       "security",
					"component": "executor",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubectl.kubernetes.io/default-container": "executor",
					},
					Labels: map[string]string{
						"app":       "security",
						"component": "executor",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "executor",
							Image: image,
							Command: []string{
								"/home/tsi/bin/executor",
							},
							Args:            args,
							VolumeMounts: []apiv1.VolumeMount{
								LogMount,
							},
							ImagePullPolicy: apiv1.PullAlways,
							SecurityContext: &apiv1.SecurityContext{
								Privileged: &trueVar,
								RunAsUser:  &sudo,
								Capabilities: &apiv1.Capabilities{
									Add: []apiv1.Capability{
										"SYS_ADMIN",
									},
								},
							},
						},
					},
					InitContainers: []apiv1.Container{
						{
							Name:  "init-mirror",
							Image: image,
							Env: []apiv1.EnvVar{
								{
									Name:  "NAME",
									Value: "executor",
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
						LogVolume,
					},
				},
			},
		},
	}

	return daemonSet
}
