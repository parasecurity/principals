package yamls

import (
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateDgaDaem(args []string, registry *string) appsv1.DaemonSet {
	var HostPathDirectoryOrCreate apiv1.HostPathType = "DirectoryOrCreate"
	// Passing path to monitor.py file to args list
	var modArgs []string = append([]string{"/tmp/monitor.py"}, args...)
	var imageDga string = *registry + ":5000/tsi-dga:v1.0.0"
	var imageMirror string = *registry + ":5000/antrea-tsi:v1.0.0"

	daemonSet := appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dga",
			Namespace: "security",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":       "security",
					"component": "dga",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubectl.kubernetes.io/default-container": "dga",
						"k8s.v1.cni.cncf.io/networks":             "macvlan-host-local",
					},
					Labels: map[string]string{
						"app":       "security",
						"component": "dga",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "dga",
							Image: imageDga,
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
							Command: []string{
								"/usr/bin/python3",
							},
							Args:            modArgs,
							ImagePullPolicy: apiv1.PullAlways,
						},
					},
					InitContainers: []apiv1.Container{
						{
							Name:  "init-mirror",
							Image: imageMirror,
							Env: []apiv1.EnvVar{
								{
									Name:  "NAME",
									Value: "dga",
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

	return daemonSet
}
