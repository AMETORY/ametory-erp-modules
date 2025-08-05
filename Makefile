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
	@curl -X POST -H "Content-Type: application/json" -d '{"username":"erpgen","content":"Just released v'$(VERSION)'","avatar_url":"https://avatars.githubusercontent.com/u/171210158"}' https://discord.com/api/webhooks/1402315010700738640/DuEN2s7g7icIN9Uw92qMSCM6lpDQNBjTPGlchvWMfezj2YLLkjBYfGM9brTH8G_L5zpD


