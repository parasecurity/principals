package yamls

import (
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateDgaForwardDaem(args []string, registry *string) appsv1.DaemonSet {
	var HostPathDirectoryOrCreate apiv1.HostPathType = "DirectoryOrCreate"
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
	var imageAntrea string = *registry + ":5000/antrea-tsi:v1.0.0"

	daemonSet := appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dga",
			Namespace: "security",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "dga",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubectl.kubernetes.io/default-container": "dga",
						"k8s.v1.cni.cncf.io/networks":             "macvlan-host-local",
					},
					Labels: map[string]string{
						"app": "dga",
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
						{
							Name:  "init-mirror",
							Image: imageAntrea,
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
