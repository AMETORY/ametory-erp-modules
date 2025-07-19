//go:generate mkdir -p templates
//go:generate cp ../../internal/generatorlib/templates/*.tmpl templates/

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/AMETORY/ametory-erp-modules/internal/generatorlib"
	"github.com/AlecAivazis/survey/v2"
	"gopkg.in/yaml.v3"
)

var version = "1.0.52"

var logo = `
MMMMMMMMMMMMMMMMWXOxoc;,............';coxOXWMMMMMMMMMMMMMMMM
MMMMMMMMMMMMWXOo:'............'...........':oOXWMMMMMMMMMMMM
MMMMMMMMMWNOl,........,'.;:'.:oc'.;,.',.......,lkXWMMMMMMMMM
MMMMMMMWXx;.....''....;,.,;'.,:;..,.'c:'':,......;dKWMMMMMMM
MMMMMMXx;....',.';;..................'..;c,.,,.....,dXMMMMMM
MMMMWO:....'..;;...........................,c,.......;kNMMMM
MMMXd'.....;,'..................................''.....oXMMM
MMXl....';;,'.........;looooooollcccc;'.........,,,,....cKMM
MXl....'',;'.........ckOOOOOOOOOkdooool,.........;:,.....cKM
No....';:;..........:xOOOkdxkOkOOxdooool'..........,;;....oX
k'...',,'..........,dOOOko,;oOOOOkxoooooc..........,,'....'k
c....';;'.........'okOOOkl.'oOOOOOkdooooo;.........'::,....:
'...',;,..........ckOOOxdl:lkOOOOOOkdooool,.........',''...'
.....,,'.........;xOOOxccoxkOOOOkOOkxdooooc'.........'''....
................'dOOOOkdlloxkOOOOOOOkxooooo:................
................lkOOOOOxl::ccldxOOOOOkdooool;...............
....,:,........:xOOOOOd;.,::,..;dOOOOkxdooool'.......,:,....
;....'........,dOOOOOkl..,::;...lOOOOOkxooooo:........,....;
d.............ckOOOOOOo;,;cl;,,;okOOOOkxdooool,............o
K:............,okkkkkko,.......'okkkkkkxdoool:............:K
WO,.............,;;;;;'.........';;;;;,,,,,''............,kW
MWk,....................................................,kWM
MMWO;..........'............................',.........;kWMM
MMMWKl.......,:,.............................,,.......c0WMMM
MMMMMNk;.....''...,;,'..................;:'.........;xNMMMMM
MMMMMMWXx;........,;;'..,;...';,...;;'..';,.......;dXWMMMMMM
MMMMMMMMWXkc'...........,,...':,...';,.........'ckXWMMMMMMMM
MMMMMMMMMMMWKxc,............................'cxKWMMMMMMMMMMM
MMMMMMMMMMMMMMWXOdc;'..................';cdOXWMMMMMMMMMMMMMM
MMMMMMMMMMMMMMMMMMWX0xl:,'........',:lx0XWMMMMMMMMMMMMMMMMMM
                                                                                          
`

var (
	coreModules = []string{
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
		"PermitHub",
	}
	// TODO: please complete this
	subModules = map[string][]string{
		"Inventory": {
			"Brand",
			"Discount",
			"Product",
			"ProductCategory",
			"ProductAttribute",
			"PriceCategory",
			"MasterProduct",
			"Purchase",
			"PurchaseReturn",
			"StockMovement",
			"StockOpname",
			"Unit",
		},
		"Company": {
			"Branch",
			"Company",
			"Organization",
			"WorkLocation",
			"Announcement",
		},
		"RBAC": {
			"Role",
		},
		"Contact": {},
		"Finance": {
			"Account",
			"Asset",
			"Bank",
			"Journal",
			"Loan",
			"Report",
			"Saving",
			"Tax",
			"Transaction",
		},
		"Cooperative": {
			"CooperativeMember",
			"CooperativeSetting",
			"LoanApplication",
			"NetSurplus",
			"Saving",
		},
		"CrowdFunding": {
			"Campaign",
			"Donation",
		},
		"CustomerRelationship": {
			"CRM",
			"Form",
			"Instagram",
			"KnowledgeBase",
			"LiveChat",
			"Telegram",
			"Ticket",
			"Whatsapp",
		},
		"Distribution": {
			"Cart",
			"Distributor",
			"Logistic",
			"Offering",
			"OrderRequest",
			"Shipping",
			"Storage",
		},
		"HRIS": {
			"Attendance",
			"Leave",
			"AttendancePolicy",
			"DeductionSetting",
			"Employee",
			"EmployeeActivity",
			"EmployeeOvertime",
			"EmployeeCashAdvance",
			"EmployeeBusinessTrip",
			"EmployeeResignation",
			"Payroll",
			"JobTitle",
			"Schedule",
			"Reimbursement",
			"EmployeeLoan",
			"WorkShift",
		},
		"Message": {
			"Chat",
			"Inbox",
		},
		"Order": {
			"Banner",
			"Merchant",
			"Payment",
			"PaymentTerm",
			"POS",
			"Promotion",
			"Sales",
			"SalesReturn",
			"Withdrawal",
		},
		"ProjectManagement": {
			"Member",
			"Project",
			"Task",
			"TaskAttribute",
			"Team",
		},
		"Tag":  {},
		"User": {},
		"PermitHub": {
			"Citizen",
			"PermitDynamicRequestData",
			"PermitUploadedDocument",
			"PermitRequest",
			"PermitApprovalLog",
			"PermitFieldDefinition",
			"PermitType",
			"PermitApprovalFlow",
			"PermitRequirement",
		},
	}
	thirdParty = []string{
		"EmailSender",
		"EmailAPIService",
		"WatzapClient",
		"WhatsmeowService",
		"Firestore",
		"FCMService",
		"RedisService",
		"KafkaService",
		"WebsocketService",
		"GoogleAPIService",
		"AppService",
	}
)

func main() {
	fmt.Println(logo)
	fmt.Printf("Welcome to ametory-erp-modules code generator! (Version: %s)\n", version)
	fmt.Println(strings.Repeat("=", 75))
	fmt.Println("Before you start, we encourage you to visit and contribute to our main repository at https://github.com/AMETORY/ametory-erp-modules")
	fmt.Println("This program will help you generate a new ametory-erp-modules project.")
	fmt.Println("Please answer the questions below to get started.")

	var answer string
	if err := survey.AskOne(&survey.Select{
		Message: "What would you like to do?",
		Options: []string{"1. Init Project", "2. Generate API", "3. Backoffice"},
		Default: "1. Init Project",
	}, &answer); err != nil {
		log.Fatal(err)
	}

	switch answer {
	case "1. Init Project":
		initProject()
	case "2. Generate API":
		generateAPI()
	case "3. Backoffice":
		initBackoffice()
	default:
		fmt.Println("Invalid answer")
	}
}

func initBackoffice() {
	projectData, err := readYamlFile()
	if err != nil {
		log.Fatalf("Failed to read project.yaml: %v", err)
		return
	}
	projectDir := projectData["project_dir"].(string)
	cloneDir := projectDir + "/backoffice"
	if err := survey.AskOne(&survey.Input{
		Message: "Enter the directory to clone the backoffice:",
		Default: cloneDir,
	}, &cloneDir); err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("git", "clone", "git@github.com:AMETORY/ametory-backoffice.git", cloneDir)
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to clone ametory-backoffice: %v, output: %s", err, string(output))
		return
	}

}
func generateAPI() {
	projectData, err := readYamlFile()
	if err != nil {
		log.Fatalf("Failed to read project.yaml: %v", err)
		return
	}
	parseCoreModules := projectData["core_modules"].([]any)
	coreModules := make([]string, len(parseCoreModules))
	for i, v := range parseCoreModules {
		coreModules[i] = v.(string)
	}
	coreModules = append(coreModules, "CUSTOM API ENDPOINT")
	var apiName string
	if err := survey.AskOne(&survey.Select{
		Message: "Enter the API name to generate:",
		Default: coreModules[0],
		Options: coreModules,
	}, &apiName); err != nil {
		log.Fatal(err)
	}

	moduleName := projectData["module_name"].(string)
	projectDir := projectData["project_dir"].(string)

	if apiName == "CUSTOM API ENDPOINT" {
		customApiName := ""
		if err := survey.AskOne(&survey.Input{
			Message: "Enter the API name to generate:",
		}, &customApiName); err != nil {
			log.Fatal(err)
		}

		subConfig := map[string]any{
			"ModuleName":       moduleName,
			"ProjectDir":       projectDir,
			"ApiName":          strings.ReplaceAll(toTitleCase(customApiName), " ", ""),
			"SnakeServiceName": strings.ReplaceAll(toSnakeCase(customApiName), " ", ""),
			"snakeApiName":     strings.ReplaceAll(toSnakeCase(customApiName), " ", ""),
			"camelApiName":     strings.ReplaceAll(toCamelCase(customApiName), " ", ""),
		}
		// fmt.Printf("Generating sub-module: %s\n", subConfig)
		// if err := generatorlib.GenerateAPI(subConfig); err != nil {
		// 	log.Fatalf("Failed to generate module %s: %v", apiName, err)
		// }
		fmt.Println("Generating API: " + customApiName)
		// utils.LogJson(subConfig)
		if err := generatorlib.GenerateAPI(subConfig); err != nil {
			log.Fatalf("Failed to generate module %s: %v", apiName, err)
		}
		return
	}

	// fmt.Printf("Generating API: %s\n", apiName)
	// fmt.Printf("Generating API: %s\n", toCamelCase(apiName))
	// fmt.Printf("Generating API: %s\n", toSnakeCase(apiName))

	listSubModules, ok := subModules[apiName]
	if ok {
		if len(listSubModules) == 0 {
			fmt.Println("No sub-modules available for this API, generating only the main API.")
			subConfig := map[string]any{
				"ModuleName":       moduleName,
				"ProjectDir":       projectDir,
				"ApiName":          apiName,
				"SnakeServiceName": toSnakeCase(apiName),
				"snakeApiName":     toSnakeCase(apiName),
				"camelApiName":     toCamelCase(apiName),
			}
			// fmt.Printf("Generating sub-module: %s\n", subConfig)
			if err := generatorlib.GenerateAPI(subConfig); err != nil {
				log.Fatalf("Failed to generate module %s: %v", apiName, err)
			}
		} else {
			var selectedSubModules []string
			if err := survey.AskOne(&survey.MultiSelect{
				Message: "Select sub-modules to generate:",
				Options: listSubModules,
			}, &selectedSubModules); err != nil {
				log.Fatal(err)
			}
			if len(selectedSubModules) > 0 {
				for _, subModule := range selectedSubModules {
					subConfig := map[string]any{
						"ModuleName":       moduleName,
						"ProjectDir":       projectDir,
						"ApiName":          subModule,
						"SnakeServiceName": toSnakeCase(apiName),
						"snakeApiName":     toSnakeCase(subModule),
						"camelApiName":     toCamelCase(subModule),
					}
					// fmt.Printf("Generating sub-module: %s\n", subConfig)
					if err := generatorlib.GenerateAPI(subConfig); err != nil {
						log.Fatalf("Failed to generate sub-module %s: %v", subModule, err)
					}
				}
			} else {
				fmt.Println("No sub-modules selected, generating only the main API.")
			}
		}
	} else {
		fmt.Println("No sub-modules available for this API, generating only the main API.")
	}
	// Placeholder for future API generation logic
}

func readYamlFile() (map[string]any, error) {
	f, err := os.ReadFile("project.yaml")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	var data map[string]any
	if err := yaml.Unmarshal(f, &data); err != nil {
		log.Fatal(err)
		return nil, err
	}

	return data, nil
}

func initProject() {
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

	var selectedCore []string
	if err := survey.AskOne(&survey.MultiSelect{
		Message: "Select core modules:",
		Options: coreModules,
	}, &selectedCore); err != nil {
		log.Fatal(err)
	}

	// Third-party selection

	var selectedThirdParty []string
	if err := survey.AskOne(&survey.MultiSelect{
		Message: "Select third-party / others integrations:",
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

	writeYAMLFile(filepath.Join(answers.ProjectDir, "project.yaml"), map[string]any{
		"generated_at":  time.Now().Format(time.RFC3339),
		"generated_by":  "ametory-erp-modules",
		"author":        "ametsuramet",
		"author email":  "ametsuramet@gmail.com",
		"version":       version,
		"module_name":   answers.ModuleName,
		"project_dir":   absPath,
		"core_modules":  selectedCore,
		"third_parties": selectedThirdParty,
	})
	fmt.Println("Project configuration saved to project.yaml")
	fmt.Println("You can now start building your ERP project!")

}
func writeYAMLFile(filePath string, data interface{}) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()

	return encoder.Encode(data)
}

func toSnakeCase(input string) string {
	if input == "RBAC" || input == "HRIS" {
		return strings.ToLower(input)
	}
	var result []rune
	for i, r := range input {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

func toTitleCase(input string) string {

	var result []rune
	for i, r := range input {
		if i == 0 || input[i-1] == '_' {
			result = append(result, unicode.ToUpper(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}
func toCamelCase(input string) string {
	if input == "RBAC" || input == "HRIS" {
		return strings.ToLower(input)
	}
	var result []rune
	nextUpper := false
	for _, r := range input {
		if r == '_' {
			nextUpper = true
		} else {
			if nextUpper {
				if len(result) == 0 {
					result = append(result, unicode.ToLower(r))
				} else {
					result = append(result, unicode.ToUpper(r))
				}
				nextUpper = false
			} else {
				result = append(result, r)
			}
		}
	}
	return strings.ToLower(string(result[0])) + string(result[1:])
}
