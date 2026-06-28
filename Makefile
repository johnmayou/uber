lint:
	@bash scripts/ci-lint

test:
	@bash scripts/ci-test

.PHONY: deploy-lint
deploy-lint:
	@kubeconform -strict \
		-schema-location default \
		-schema-location 'https://raw.githubusercontent.com/datreeio/CRDs-catalog/main/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json' \
		deploy/manifests/

.PHONY: argo
argo:
	@kubectl port-forward svc/argocd-server -n argocd 8080:443

.PHONY: argo-password
argo-password:
	@kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d | pbcopy
	@echo "copied to clipboard"