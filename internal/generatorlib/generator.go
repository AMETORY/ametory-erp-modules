package generatorlib

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"

	"github.com/schollz/progressbar/v3"
)

//go:embed templates/*
var templateFS embed.FS

func loadTemplate(name string) (*template.Template, error) {
	content, err := fs.ReadFile(templateFS, "templates/"+name)
	if err != nil {
		return nil, err
	}

	tmpl := template.New(name).Funcs(template.FuncMap{
		"contains": func(list []string, item string) bool {
			for _, s := range list {
				if s == item {
					return true
				}
			}
			return false
		},
	})

	return tmpl.Parse(string(content))
}

type ProjectConfig struct {
	ModuleName   string
	ProjectDir   string
	CoreModules  []string
	ThirdParties []string
}

func GenerateProject(config ProjectConfig) error {
	// Create project directory
	if err := os.MkdirAll(config.ProjectDir, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Initialize progress bar
	bar := progressbar.NewOptions(4,
		progressbar.OptionSetDescription("Generating project..."),
	)

	// 1. Create go.mod file
	if err := createGoMod(config); err != nil {
		return err
	}
	bar.Add(1)

	// 2. Generate main.go
	if err := generateFromTemplate("main.go.tmpl",
		filepath.Join(config.ProjectDir, "main.go"), config); err != nil {
		return err
	}
	bar.Add(1)

	// 3. Generate config file
	if err := generateFromTemplate("config.yaml.tmpl",
		filepath.Join(config.ProjectDir, "config.yaml"), config); err != nil {
		return err
	}
	bar.Add(1)

	// 4. Create other necessary files
	if err := createGitIgnore(config.ProjectDir); err != nil {
		return err
	}
	bar.Add(1)

	// 5. Create directories for api handler
	if err := generateFromTemplate("apihandler.go.tmpl",
		filepath.Join(config.ProjectDir, "api", "handler", "auth.go"), config); err != nil {
		return err
	}
	bar.Add(1)

	// 6. Create directories for api router
	if err := generateFromTemplate("apirouter.go.tmpl",
		filepath.Join(config.ProjectDir, "api", "router", "auth.go"), config); err != nil {
		return err
	}
	bar.Add(1)

	// 7. Create directories for config
	if err := generateFromTemplate("config.go.tmpl",
		filepath.Join(config.ProjectDir, "config", "config.go"), config); err != nil {
		return err
	}
	bar.Add(1)

	if err := generateFromTemplate("database.go.tmpl",
		filepath.Join(config.ProjectDir, "config", "database.go"), config); err != nil {
		return err
	}
	bar.Add(1)

	if err := generateFromTemplate("email.go.tmpl",
		filepath.Join(config.ProjectDir, "config", "email.go"), config); err != nil {
		return err
	}
	bar.Add(1)

	if err := generateFromTemplate("server.go.tmpl",
		filepath.Join(config.ProjectDir, "config", "server.go"), config); err != nil {
		return err
	}
	bar.Add(1)

	return nil
}

func createGoMod(config ProjectConfig) error {
	content := fmt.Sprintf(`module %s

go 1.21

require (
	github.com/AMETORY/ametory-erp-modules v1.0.1
)`, config.ModuleName)

	return os.WriteFile(
		filepath.Join(config.ProjectDir, "go.mod"),
		[]byte(content),
		0644,
	)
}

func generateFromTemplate(templateName, outputPath string, data ProjectConfig) error {
	// Buat template dengan fungsi custom

	tmpl, err := loadTemplate(templateName)
	if err != nil {
		return err
	}

	// Buat direktori jika belum ada
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, data)

}

func createGitIgnore(projectDir string) error {
	content := `bin/
obj/
.vscode/
.idea/
*.exe
*.test
*.out
`
	return os.WriteFile(
		filepath.Join(projectDir, ".gitignore"),
		[]byte(content),
		0644,
	)
}
