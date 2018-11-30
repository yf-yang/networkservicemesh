docker save -o scripts/vagrant/images/skydive.tar skydive/skydive:devel
make k8s-skydive-load-images
kubectl delete deployment skydive-analyzer
kubectl create -f scripts/vagrant/skydive.yaml
