KAGENT_NAMESPACE ?= kagent
KAGENT_PROVIDER ?= ollama
VERSION ?= 0.8.3
KMCP_VERSION ?= 0.2.8

.PHONY: helm-version
helm-version:
	@echo "Generating Chart.yaml files (VERSION=$(VERSION), KMCP_VERSION=$(KMCP_VERSION))..."
	@for f in kagent-crds/Chart-template.yaml kagent/Chart-template.yaml; do \
		VERSION=$(VERSION) KMCP_VERSION=$(KMCP_VERSION) envsubst < $$f > $$(dirname $$f)/Chart.yaml; \
	done
	@for f in agents/*/Chart-template.yaml tools/*/Chart-template.yaml; do \
		VERSION=$(VERSION) envsubst < $$f > $$(dirname $$f)/Chart.yaml; \
	done
	helm dependency update kagent-crds
	helm dependency update kagent

.PHONY: deploy-ollama
deploy-ollama:
	kubectl apply -f ollama.yaml
	@echo "Waiting for Ollama pod to be ready..."
	kubectl wait --for=condition=ready pod -l app=ollama -n $(KAGENT_NAMESPACE) --timeout=300s
	@echo "Pulling llama3.1:8b model..."
	kubectl exec -n $(KAGENT_NAMESPACE) deployment/ollama -- ollama pull llama3.1:8b

.PHONY: helm-install
helm-install: helm-version
	helm upgrade --install kagent-crds ./kagent-crds/ \
		--namespace $(KAGENT_NAMESPACE) \
		--create-namespace
	helm upgrade --install kagent ./kagent/ \
		--namespace $(KAGENT_NAMESPACE) \
		--create-namespace \
		-f values-arohcp.yaml

.PHONY: helm-uninstall
helm-uninstall:
	helm uninstall kagent -n $(KAGENT_NAMESPACE) --wait --timeout=5m 2>/dev/null || true
	helm uninstall kagent-crds -n $(KAGENT_NAMESPACE) --wait --timeout=2m 2>/dev/null || true
	kubectl delete namespace $(KAGENT_NAMESPACE) --ignore-not-found=true

.PHONY: port-forward-ui
port-forward-ui:
	@echo "kagent UI available at http://localhost:8001"
	kubectl port-forward svc/kagent-ui -n $(KAGENT_NAMESPACE) 8001:8080

.PHONY: clean
clean:
	rm -f kagent-crds/Chart.yaml kagent-crds/Chart.lock
	rm -f kagent/Chart.yaml kagent/Chart.lock
	rm -rf kagent-crds/charts/ kagent/charts/
	rm -f agents/*/Chart.yaml tools/*/Chart.yaml
