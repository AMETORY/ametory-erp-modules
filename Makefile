.PHONY: tag push-tag

VERSION ?= $(shell git describe --tags --abbrev=0)

tag:
	sed -i '' "s|version = \".*\"|version = \"$(VERSION)\"|g" cmd/erpgen/main.go
	sed -i '' "s|version = \".*\"|version = \"$(VERSION)\"|g" internal/generatorlib/generator.go
	git add .
	git commit -m "release v$(VERSION)"
	git push origin main
	git tag v$(VERSION)


# example: make release VERSION=1.0.1
release: tag
	git push origin v$(VERSION)


