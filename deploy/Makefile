.DEFAULT_GOAL := help

k8s-domain = example.com
ingress-gw = $(shell kubectl -n istio-system get svc istio-ingressgateway \
    -o jsonpath="{.status.loadBalancer.ingress[0].ip}")


.PHONY: help
help:
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	    awk 'BEGIN {FS = ":.*?## "}; \
		{printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: deploy-to-cf-with-buildpack
deploy-to-cf-with-buildpack: ## Deploy t2m to cf using go buildpack
	cf push -f cf/manifest-bp.yml
	@echo To call application use:
	@echo curl http://t2m.pub.cf.frnksgr.net

.PHONY: deploy-to-cf-with-image
deploy-to-cf-with-image: ## Deploy docker image frnksgr/t2m to cf
	cf push -f cf/manifest.yml
	@echo To call application use:
	@echo curl http://t2m.pub.cf.frnksgr.net

.PHONY: deploy-to-cf
deploy-to-cf: deploy-to-cf-with-image ## Deploy to cf

.PHONY: delete-from-cf
delete-from-cf: ## Delete t2m deployment on CF
	cf delete -f -r t2m

.PHONY: deploy-to-k8s
deploy-to-k8s: ## Deploy t2m as k8s service with ingress
	kubectl apply -f k8s/service-with-ingress.yml
	@echo To call application use:
	@echo curl -H \"Host: t2m.default.$(k8s-domain)\" http://$(ingress-gw)/	

.PHONY: delete-from-k8s
delete-from-k8s: ## Delete any t2m related deployments on k8s cluster
	@echo not implemented yet
	#kubectl delete delete -f k8s/service-with-ingress.yml --ignore-not-found=true
