.PHONY: lint
lint:
	make -C api lint
	make -C web lint

.PHONY: test
test:
	make -C api test
	make -C web test

.PHONY: deploy-lint
deploy-lint:
	@kubeconform -strict \
		-schema-location default \
		-schema-location 'https://raw.githubusercontent.com/datreeio/CRDs-catalog/main/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json' \
		deploy/manifests/
