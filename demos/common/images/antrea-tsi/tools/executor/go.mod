module executor

go 1.13

replace logging => ../../../logging

require (
	github.com/containerd/cgroups v1.0.1
	github.com/google/gopacket v1.1.19
	github.com/opencontainers/runtime-spec v1.0.2
	logging v0.0.0-00010101000000-000000000000
)
