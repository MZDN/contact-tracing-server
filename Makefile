.PHONY: coepi

GOBIN = $(shell pwd)/bin
GO ?= latest
DEV=mm

findmypk:
	go build -o bin/findmypk
	@echo "Done building FindMyPk!  Run \"$(GOBIN)/findmypk\" to launch findmypk."

docker:
	docker build --force-rm -t gcr.io/us-west1-wlk/wolkinc/findmypk-$(DEV) .
	gcloud docker -- push gcr.io/us-west1-wlk/wolkinc/findmypk-$(DEV):latest

createpvcs:
	kubectl apply -f ./build/yaml/pvc.yaml

getpvcs:
	kubectl get pvc findmypk

deletepvcs:
	kubectl delete pvc findmypk

createservices:
	kubectl apply -f ./build/yaml/service.yaml

getservices:
	kubectl get services findmypk

deleteservices:
	kubectl delete service findmypk

createpods:
	kubectl apply -f ./build/yaml/deploy.yaml
	kubectl autoscale findmypk --max 6 --min 3 --cpu-percent 50

getpods:
	kubectl get pods

getpod:
	kubectl describe pod findmypk

deletepods:
	kubectl delete deployment findmypk

restartpod:
	kubectl delete -f ./build/yaml/$(POD)-cloudstore.yml --namespace=$(NAMESPACE) --kubeconfig=/root/.kube/$(CLUSTER)
	kubectl create -f ./build/yaml/$(POD)-cloudstore.yml --namespace=$(NAMESPACE) --kubeconfig=/root/.kube/$(CLUSTER)

pods:
	-make deletepods
	make createpods

sshpod:
	kubectl exec -it `kubectl get pods | grep findmypk | awk '{print $$1}'` -- bash
