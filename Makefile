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
