.PHONY: coepi

GOBIN = $(shell pwd)/bin
GO ?= latest

findmypk:
		go build -o bin/findmypk
		@echo "Done building FindMyPk!  Run \"$(GOBIN)/findmypk\" to launch findmypk."

docker:
	docker build --force-rm -t gcr.io/us-west1-wlk/wolkinc/findmypk-mm .
	gcloud docker -- push gcr.io/us-west1-wlk/wolkinc/findmypk-mm:latest

createcluster:
	-make createips
	-make stage$(STAGE)
	-gcloud container clusters create $(CLUSTER) --num-nodes=3 --image-type "UBUNTU" --disk-size=100 --disk-type=pd-standard --machine-type=n1-standard-4 --enable-autorepair --no-enable-cloud-logging --no-enable-cloud-monitoring --issue-client-certificate --enable-ip-alias --zone=$(ZONE) --scopes=https://www.googleapis.com/auth/cloud-platform  --tags=allowall,http-server,https-server
	-make getconfig
	-make createnamespace
	-make createservices
	-make createpvcs
	-make createpods
	-cat cheatsheet$(STAGE).txt
	-watch make getservices

getconfig:
	-gcloud container clusters get-credentials $(CLUSTER) --zone $(ZONE) --project $(PROJECT); 'yes' | cp -rf /root/.kube/config /root/.kube/$(CLUSTER); > /root/.kube/config

deletecluster:
	-make deletepods
	-make deletepvcs
	-make deleteservices
	-deletedns $(STAGE)
	-deleteforwardingrules $(REGION)
	-deletetargetpools $(REGION)
	-make deleteips
	-gcloud container clusters delete $(CLUSTER) --zone $(ZONE) --project $(PROJECT) --quiet

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

createips:
	# get the ips from google
	-for i in `seq 0 $(LAST)` ; do gcloud beta compute --project=$(PROJECT) addresses create cw$$i-ip --region=$(REGION) --network-tier=PREMIUM; done
	# update servers table with the new ips
	-createip $(STAGE)
	# setup google cloud DNS entries like w0-6.wolk.com
	-updatedns $(STAGE)

getips:
	gcloud beta compute --project=$(PROJECT) addresses list | grep $(CLUSTER)

deleteips:
	for i in `seq 0 $(LAST)` ; do \
	  gcloud -q beta compute --project=$(PROJECT) addresses delete cw$$i-ip --region $(REGION); \
	done

createnamespace:
	kubectl create -f ./build/conf/namespace-$(NAMESPACE).json --kubeconfig=/root/.kube/cw

getnamespace:
	kubectl get namespaces --show-labels  --kubeconfig=/root/.kube/cw
