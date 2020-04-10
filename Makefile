.PHONY: coepi

GOBIN = $(shell pwd)/bin
GO ?= latest

contact-tracing:
	go build -o bin/contact-tracing
	@echo "Done building Contact Tracing!  Run \"$(GOBIN)/contact-tracing\" to launch contact-tracing."

docker:
	docker build --force-rm -t gcr.io/us-west1-wlk/wolkinc/contact-tracing .
	gcloud docker -- push gcr.io/us-west1-wlk/wolkinc/contact-tracing:latest

createpvcs:
	kubectl apply -f ./build/yaml/pvc.yaml

getpvcs:
	kubectl get pvc contact-tracing

deletepvcs:
	kubectl delete pvc contact-tracing

createservices:
	kubectl apply -f ./build/yaml/service.yaml

getservices:
	kubectl get services contact-tracing

deleteservices:
	kubectl delete service contact-tracing

createpods:
	kubectl apply -f ./build/yaml/deploy.yaml
	kubectl apply -f ./build/yaml/autoscale.yaml

getpods:
	kubectl get pods

getpod:
	kubectl describe pod contact-tracing

gethpa:
	kubectl get hpa contact-tracing

deletepods:
	- kubectl delete hpa contact-tracing
	kubectl delete deployment contact-tracing

restartpod:
	kubectl delete -f ./build/yaml/$(POD)-cloudstore.yml --namespace=$(NAMESPACE) --kubeconfig=/root/.kube/$(CLUSTER)
	kubectl create -f ./build/yaml/$(POD)-cloudstore.yml --namespace=$(NAMESPACE) --kubeconfig=/root/.kube/$(CLUSTER)

pods:
	-make deletepods
	make createpods

sshpod:
	kubectl exec -it `kubectl get pods | grep contact-tracing | awk '{print $$1}'` -- bash
