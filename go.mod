module containers-migration-factory

go 1.16

require (
	github.com/aws/aws-sdk-go v1.43.16
	github.com/bigkevmcd/go-configparser v0.0.0-20210106142102-909504547ead
	github.com/ghodss/yaml v1.0.0
	github.com/gofrs/flock v0.8.1
	github.com/pkg/errors v0.9.1
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	helm.sh/helm/v3 v3.14.1
	k8s.io/api v0.29.0
	k8s.io/apimachinery v0.29.0
	k8s.io/client-go v0.29.0

)

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
	github.com/opencontainers/image-spec => github.com/opencontainers/image-spec v1.0.2
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.1.2
)
