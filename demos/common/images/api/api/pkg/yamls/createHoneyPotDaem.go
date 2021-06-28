package yamls

import (
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateHoneypotDaem(args []string) appsv1.DaemonSet {
	// Copying IP and mac address to shared drive
	// Each service that wants access to those values
	// just needs to mount the shared drive
	var modArgs []string = append([]string{
		"cp /sys/class/net/eth0/address /home/net/dionaea_mac &&",
		"hostname -i > /home/net/dionaea_ip &&",
		"/opt/dionaea/bin/dionaea "},
		args...)
	modArgs = append(modArgs, []string{"&& sleep infinity"}...)
	var stringArgs string = strings.Join(modArgs[:], " ")

	daemonSet := appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "honeypot",
			Namespace: "security",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "honeypot",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubectl.kubernetes.io/default-container": "honeypot",
					},
					Labels: map[string]string{
						"app": "honeypot",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "honeypot",
							Image: "147.27.39.116:5000/tsi-honeypot:v1.0.0",
							Command: []string{
								"bash",
								"-c",
							},
							Args: []string{
								stringArgs,
							},
							ImagePullPolicy: apiv1.PullAlways,
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "dionaea",
									MountPath: "/home/net",
								},
							},
							Ports: []apiv1.ContainerPort{
								{
									Name:          "connect",
									ContainerPort: 80,
								},
							},
							SecurityContext: &apiv1.SecurityContext{
								Capabilities: &apiv1.Capabilities{
									Add: []apiv1.Capability{
										"NET_ADMIN",
									},
								},
							},
							Env: []apiv1.EnvVar{
								{
									Name: "NODE_NAME",
									ValueFrom: &apiv1.EnvVarSource{
										FieldRef: &apiv1.ObjectFieldSelector{
											FieldPath: "spec.nodeName",
										},
									},
								},
							},
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name: "dionaea",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/mnt/data/security_vol",
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
