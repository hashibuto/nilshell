GIT_REPO := https://github.com/hashibuto/nilshell
GIT_DEFAULT_BRANCH := master
GOMARKDOC_OPTIONS := --repository.url $(GIT_REPO) --repository.path / --repository.default-branch $(GIT_DEFAULT_BRANCH) --output ./docs/doc.md .

.PHONY: docs

docs: install-tools
	gomarkdoc $(GOMARKDOC_OPTIONS)

check-docs: install-tools
	gomarkdoc --check $(GOMARKDOC_OPTIONS)

install-tools:
	go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest

test:
	go test
