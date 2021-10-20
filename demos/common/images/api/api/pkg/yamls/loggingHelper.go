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
