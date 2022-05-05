module api

go 1.13

replace logging => ../../logging

require (
	github.com/imdario/mergo v0.3.12 // indirect
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba // indirect
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
	k8s.io/utils v0.0.0-20210517184530-5a248b5acedc // indirect
)
