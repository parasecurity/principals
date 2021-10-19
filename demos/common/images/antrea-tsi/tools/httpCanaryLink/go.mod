module canaryLink

go 1.13

require github.com/antrea-io/antrea v0.0.0-00010101000000-000000000000

replace (
	logging => ../../../logging
	github.com/antrea-io/antrea => antrea.io/antrea v1.1.0
	github.com/contiv/ofnet => github.com/contiv/ofnet v0.0.0-20180104211757-c080e5b6e9be
)
