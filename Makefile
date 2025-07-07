.PHONY: tag push-tag

VERSION ?= $(shell git describe --tags --abbrev=0)

tag:
	git tag $(VERSION)


# example: make release VERSION=1.0.1
release: tag
	sed -i '' "s|version = \".*\"|version = \"v$(VERSION)\"|g" cmd/erpgen/main.go
	sed -i '' "s|version = \".*\"|version = \"v$(VERSION)\"|g" internal/generatorlib/generator.go
	git add cmd/erpgen/main.go
	git add internal/generatorlib/generator.go
	git commit -m "release $(VERSION)"
	git push origin $(VERSION)


