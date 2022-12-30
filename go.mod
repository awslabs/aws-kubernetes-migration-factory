module containers-migration-factory

go 1.16

require (
	github.com/Shopify/logrus-bugsnag v0.0.0-20171204204709-577dee27f20d // indirect
	github.com/aws/aws-sdk-go v1.33.0
	github.com/bigkevmcd/go-configparser v0.0.0-20210106142102-909504547ead
	github.com/bshuster-repo/logrus-logstash-hook v1.1.0 // indirect
	github.com/bugsnag/bugsnag-go v2.2.0+incompatible // indirect
	github.com/bugsnag/panicwrap v1.3.4 // indirect
	github.com/containerd/containerd v1.4.13 // indirect
	github.com/docker/distribution v2.8.0-beta.1+incompatible // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/garyburd/redigo v1.6.4 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/gofrs/flock v0.8.0
	github.com/gorilla/handlers v1.5.1 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/pkg/errors v0.9.1
	github.com/yvasiyarov/go-metrics v0.0.0-20150112132944-c25f46c4b940 // indirect
	github.com/yvasiyarov/gorelic v0.0.7 // indirect
	github.com/yvasiyarov/newrelic_platform_go v0.0.0-20160601141957-9c099fbc30e9 // indirect
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
