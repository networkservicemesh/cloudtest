.PHONY: yaml-lint
yaml-lint:
	@yamllint -c .yamllint.yml --strict .