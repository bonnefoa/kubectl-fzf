package k8sresources

import "encoding/gob"

func init() {
	gob.Register(APIResource{})
	gob.Register(&ConfigMap{})
	gob.Register(&CronJob{})
	gob.Register(&DaemonSet{})
	gob.Register(&Deployment{})
	gob.Register(&Endpoints{})
	gob.Register(&Hpa{})
	gob.Register(&Ingress{})
	gob.Register(&Job{})
	gob.Register(&Namespace{})
	gob.Register(&Node{})
	gob.Register(&Pod{})
	gob.Register(&PersistentVolume{})
	gob.Register(&PersistentVolumeClaim{})
	gob.Register(&ReplicaSet{})
	gob.Register(&Secret{})
	gob.Register(&Service{})
	gob.Register(&ServiceAccount{})
	gob.Register(&StatefulSet{})
}
