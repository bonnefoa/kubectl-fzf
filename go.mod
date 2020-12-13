module kubectlfzf

go 1.15

replace cmd/cache_builder => ./cache_builder

require (
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sevlyar/go-daemon v0.1.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	k8s.io/api v0.20.0
	k8s.io/apimachinery v0.20.0
	k8s.io/client-go v0.20.0
)
