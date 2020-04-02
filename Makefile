.PHONY: coepi

GOBIN = $(shell pwd)/bin
GO ?= latest

cen:
		go build -o bin/cen
		@echo "Done building cen.  Run \"$(GOBIN)/cen\" to launch cen."

docker: 
	docker build --force-rm -t gcr.io/us-west1-wlk/wolkinc/cen_mm build
	gcloud docker -- push gcr.io/us-west1-wlk/wolkinc/cen_mm:latest

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
	for i in `seq 0 $(LAST)` ; do \
	  kubectl create -f ./build/yaml/w$$i-$(STAGE)-pvc.yml --namespace=$(NAMESPACE) --kubeconfig=/root/.kube/$(CLUSTER); \
        done

getpvcs:
	kubectl get pvc --namespace=$(NAMESPACE) --kubeconfig=/root/.kube/$(CLUSTER)

deletepvcs:
	for i in `seq 0 $(LAST)` ; do \
	  kubectl delete -f ./build/yaml/w$$i-$(STAGE)-pvc.yml --namespace=$(NAMESPACE) --kubeconfig=/root/.kube/$(CLUSTER); \
	done

createservices:
	for i in `seq 0 $(LAST)` ; do \
	  kubectl create -f ./build/yaml/w$$i-$(STAGE)-service.yml --namespace=$(NAMESPACE) --kubeconfig=/root/.kube/$(CLUSTER); \
	done

getservices:
	kubectl get services --namespace=$(NAMESPACE) --kubeconfig=/root/.kube/$(CLUSTER)

deleteservices:
	for i in `seq 0 $(LAST)` ; do \
	  kubectl delete -f ./build/yaml/w$$i-$(STAGE)-service.yml --kubeconfig=/root/.kube/$(CLUSTER); \
	done

createpods:
	for i in `seq 0 $(LAST)` ; do \
	  kubectl create -f ./build/yaml/w$$i-$(STAGE)-cloudstore.yml --namespace=$(NAMESPACE) --kubeconfig=/root/.kube/$(CLUSTER); \
	done

getpods:
	kubectl get pods --namespace=$(NAMESPACE)  --kubeconfig=/root/.kube/$(CLUSTER)

getpod:
	kubectl describe pod $(POD) --namespace=$(NAMESPACE) --kubeconfig=/root/.kube/$(CLUSTER)

deletepods:
	for i in `seq 0 $(LAST)` ; do \
	  kubectl delete -f ./build/yaml/w$$i-$(STAGE)-cloudstore.yml --namespace=$(NAMESPACE) --kubeconfig=/root/.kube/$(CLUSTER); \
	done

restartpod:
	kubectl delete -f ./build/yaml/$(POD)-cloudstore.yml --namespace=$(NAMESPACE) --kubeconfig=/root/.kube/$(CLUSTER)
	kubectl create -f ./build/yaml/$(POD)-cloudstore.yml --namespace=$(NAMESPACE) --kubeconfig=/root/.kube/$(CLUSTER)

pods:
	-make deletepods
	make createpods

sshpod:
	kubectl exec --kubeconfig=/root/.kube/$(CLUSTER) --namespace=$(NAMESPACE) -it `kubectl get pods --namespace=$(NAMESPACE)  --kubeconfig=/root/.kube/$(CLUSTER) | grep $(POD) | awk '{print $$1}'` -- bash

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

