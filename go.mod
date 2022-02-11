module containers-migration-factory

go 1.16

require (
	github.com/aws/aws-sdk-go v1.27.0
	github.com/bigkevmcd/go-configparser v0.0.0-20210106142102-909504547ead
	github.com/ghodss/yaml v1.0.0
	github.com/gofrs/flock v0.8.0
	github.com/pkg/errors v0.9.1
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
	helm.sh/helm/v3 v3.5.3
	k8s.io/api v0.20.5
	k8s.io/apimachinery v0.20.5
	k8s.io/client-go v0.20.5

)

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
)
