package yamls

import (
	apiv1 "k8s.io/api/core/v1"
)

var HostPathDirectoryOrCreate apiv1.HostPathType = "DirectoryOrCreate"

var LogMount apiv1.VolumeMount = apiv1.VolumeMount{
	Name:	"logsock",
	MountPath: "/tmp",
}

var LogVolume apiv1.Volume = apiv1.Volume{
	Name: "logsock",
	VolumeSource: apiv1.VolumeSource{
		HostPath: &apiv1.HostPathVolumeSource{
			Path: "/tmp",
			Type: &HostPathDirectoryOrCreate,
		},
	},
}

var RunAntreaVolume apiv1.Volume = apiv1.Volume{
	Name: "host-var-run-antrea",
	VolumeSource: apiv1.VolumeSource{
		HostPath: &apiv1.HostPathVolumeSource{
			Path: "/var/run/antrea",
			Type: &HostPathDirectoryOrCreate,
		},
	},
}

func mirrorContainter(name string, registry *string) apiv1.Container {
	var image string = *registry + ":5000/antrea-tsi:v1.0.1"
	mc := apiv1.Container {
		Name:  "init-mirror",
		Image: image,
		Env: []apiv1.EnvVar{
			{
				Name:  "NAME",
				Value: name,
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
	}
	return mc
}
