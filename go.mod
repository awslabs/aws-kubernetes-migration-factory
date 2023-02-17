module containers-migration-factory

go 1.16

require (
	github.com/Microsoft/go-winio v0.4.16 // indirect
	github.com/Microsoft/hcsshim v0.8.14 // indirect
	github.com/Shopify/logrus-bugsnag v0.0.0-20230117174420-439a4b8ba167 // indirect
	github.com/aws/aws-sdk-go v1.43.16
	github.com/bigkevmcd/go-configparser v0.0.0-20210106142102-909504547ead
	github.com/bshuster-repo/logrus-logstash-hook v1.1.0 // indirect
	github.com/bugsnag/bugsnag-go v2.2.0+incompatible // indirect
	github.com/containerd/continuity v0.0.0-20201208142359-180525291bb7 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/emicklei/go-restful v2.16.0+incompatible // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-gorp/gorp/v3 v3.1.0 // indirect
	github.com/gofrs/flock v0.8.1
	github.com/opencontainers/runc v0.1.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/yvasiyarov/go-metrics v0.0.0-20150112132944-c25f46c4b940 // indirect
	github.com/yvasiyarov/gorelic v0.0.7 // indirect
	github.com/yvasiyarov/newrelic_platform_go v0.0.0-20160601141957-9c099fbc30e9 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	helm.sh/helm/v3 v3.11.1
	k8s.io/api v0.26.0
	k8s.io/apimachinery v0.26.0
	k8s.io/client-go v0.26.0
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.13
	github.com/docker/distribution => github.com/docker/distribution v2.8.0-beta.1+incompatible
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
	github.com/opencontainers/image-spec => github.com/opencontainers/image-spec v1.0.2
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.3
)
