//go:generate mkdir -p templates
//go:generate cp ../../internal/generatorlib/templates/*.tmpl templates/

package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/AMETORY/ametory-erp-modules/internal/generatorlib"
	"github.com/AlecAivazis/survey/v2"
)

func main() {
	fmt.Println("Welcome to ametory-erp-modules code generator!")
	fmt.Println("This program will help you generate a new ametory-erp-modules project.")
	fmt.Println("Please answer the questions below to get started.")
	// Basic project info
	qs := []*survey.Question{
		{
			Name: "moduleName",
			Prompt: &survey.Input{
				Message: "Go module name:",
				Help:    "e.g., github.com/username/my-erp-project",
				Default: "github.com/example/erp-project",
			},
			Validate: survey.Required,
		},
		{
			Name: "projectDir",
			Prompt: &survey.Input{
				Message: "Project directory:",
				Default: "./my-erp-project",
			},
		},
	}

	answers := struct {
		ModuleName string
		ProjectDir string
	}{}

	if err := survey.Ask(qs, &answers); err != nil {
		log.Fatal(err)
	}

	// Module selection
	coreModules := []string{
		"Auth",
		"AdminAuth",
		"RBAC",
		"Inventory",
		"Manufacture",
		"Company",
		"Contact",
		"Finance",
		"Cooperative",
		"Order",
		"Logistic",
		"AuditTrail",
		"Distribution",
		"CustomerRelationship",
		"File",
		"Medical",
		"IndonesiaReg",
		"User",
		"ContentManagement",
		"Tag",
		"Message",
		"ProjectManagement",
		"CrowdFunding",
		"Notification",
		"HRIS",
	}

	var selectedCore []string
	if err := survey.AskOne(&survey.MultiSelect{
		Message: "Select core modules:",
		Options: coreModules,
	}, &selectedCore); err != nil {
		log.Fatal(err)
	}

	// Third-party selection
	thirdParty := []string{
		"EmailSender",
		"EmailAPIService",
		"WatzapClient",
		"WhatsmeowService",
		"Firestore",
		"FCMService",
		"AppService",
	}

	var selectedThirdParty []string
	if err := survey.AskOne(&survey.MultiSelect{
		Message: "Select third-party integrations:",
		Options: thirdParty,
	}, &selectedThirdParty); err != nil {
		log.Fatal(err)
	}

	// Confirm
	confirm := false
	if err := survey.AskOne(&survey.Confirm{
		Message: fmt.Sprintf("Create project at %s with module name %s?",
			answers.ProjectDir, answers.ModuleName),
	}, &confirm); err != nil || !confirm {
		fmt.Println("Operation cancelled")
		return
	}

	// Generate project
	absPath, _ := filepath.Abs(answers.ProjectDir)
	config := generatorlib.ProjectConfig{
		ModuleName:   answers.ModuleName,
		ProjectDir:   absPath,
		CoreModules:  selectedCore,
		ThirdParties: selectedThirdParty,
	}

	if err := generatorlib.GenerateProject(config); err != nil {
		log.Fatalf("Failed to generate project: %v", err)
	}

	fmt.Printf("\n✅ Successfully created project at: %s\n", absPath)
	fmt.Println("Next steps:")
	fmt.Printf("1. cd %s\n", answers.ProjectDir)
	fmt.Println("2. go mod tidy")
	fmt.Println("3. go run main.go")
	fmt.Println("4. check third parties parameters and Customize your application as needed. You can add more modules or third-party integrations later.")
}
