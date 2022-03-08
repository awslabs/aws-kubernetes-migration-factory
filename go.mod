module containers-migration-factory

go 1.16

require (
	github.com/aws/aws-sdk-go v1.27.0
	github.com/bigkevmcd/go-configparser v0.0.0-20210106142102-909504547ead
	github.com/containerd/containerd v1.4.13 // indirect
	github.com/docker/distribution v2.8.0-beta.1+incompatible // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/gofrs/flock v0.8.0
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/pkg/errors v0.9.1
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
	helm.sh/helm/v3 v3.6.1
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v0.21.0
	rsc.io/letsencrypt v0.0.3 // indirect
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.13
	github.com/docker/distribution => github.com/docker/distribution v2.8.0-beta.1+incompatible
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
	github.com/opencontainers/image-spec => github.com/opencontainers/image-spec v1.0.2
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.3
)
