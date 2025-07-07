.PHONY: tag push-tag

VERSION ?= $(shell git describe --tags --abbrev=0)

# example: make release VERSION=1.0.1
release:
	@echo "Updating version in cmd/erpgen/main.go"
	@sed -i '' "s|version = \".*\"|version = \"$(VERSION)\"|g" cmd/erpgen/main.go
	@echo "Updating version in internal/generatorlib/generator.go"
	@sed -i '' "s|version = \".*\"|version = \"$(VERSION)\"|g" internal/generatorlib/generator.go
	@echo "Staging changes"
	@git add .
	@echo "Committing changes with message: release v$(VERSION)"
	@git commit -m "release v$(VERSION)"
	@echo "Pushing changes to origin main"
	@git push origin main
	@echo "Creating tag v$(VERSION)"
	@git tag v$(VERSION)
	@echo "Pushing tag v$(VERSION) to origin"
	@git push origin v$(VERSION)


