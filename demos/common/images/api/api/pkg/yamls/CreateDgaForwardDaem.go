package yamls

import (
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateDgaForwardDaem(args []string, registry *string) appsv1.DaemonSet {
	// Passing path to monitor.py file to args list
	// And declare HONEYPOT_MAC and HONEYPOT_IP address
	var modArgs []string = append([]string{
		"export HONEYPOT_MAC=$(cat /home/dionaea/dionaea_mac) &&",
		"export HONEYPOT_IP=$(cat /home/dionaea/dionaea_ip) &&",
		"/usr/bin/python3 /tmp/monitor.py "},
		args...)
	modArgs = append(modArgs, []string{` -arg="{\""honeypot_ip\"": \""$HONEYPOT_IP\"", \""honeypot_mac\"": \""$HONEYPOT_MAC\""}"`}...)
	var stringArgs string = strings.Join(modArgs[:], " ")
	var imageDga string = *registry + ":5000/tsi-dga:v1.0.0"
	var imageAntrea string = *registry + antrea_image

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
									MountPath: "/home/dionaea",
								},
							},
						},
					},
					InitContainers: []apiv1.Container{
						{
							Name:  "init-addr",
							Image: imageAntrea,
							Command: []string{
								"bash",
							},
							Args: []string{
								"-c",
								"while [ ! -f /home/dionaea/dionaea_mac ];do sleep 2;done && while [ ! -f /home/dionaea/dionaea_ip ];do sleep 2;done",
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "dionaea",
									MountPath: "/home/dionaea",
								},
							},
						},
						mirrorContainter("dga", registry),
					},
					Volumes: []apiv1.Volume{
						RunAntreaVolume,
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
