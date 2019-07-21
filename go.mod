module github.com/bonnefoa/kubectl-fzf

replace github.com/bonnefoa/kubectl-fzf/cmd/cache_builder => ./cache_builder

require (
	cloud.google.com/go v0.39.0 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/gofuzz v1.0.0 // indirect
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20190212212710-3befbb6ad0cc // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/json-iterator/go v1.1.6 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v0.0.0-20180320133207-05fbef0ca5da // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.3.0
	golang.org/x/net v0.0.0-20190606173856-1492cefac77f // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/sys v0.0.0-20190606203320-7fc4e5ec1444 // indirect
	google.golang.org/appengine v1.6.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/api v0.0.0-20181130031204-d04500c8c3dd
	k8s.io/apimachinery v0.0.0-20181215012845-4d029f033399
	k8s.io/client-go v10.0.0+incompatible
	k8s.io/klog v0.3.0 // indirect
	sigs.k8s.io/yaml v1.1.0 // indirect
)
