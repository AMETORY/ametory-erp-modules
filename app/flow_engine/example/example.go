package main

import (
	"fmt"

	. "github.com/AMETORY/ametory-erp-modules/app/flow_engine"
)

// 11. Contoh Function yang akan diregistrasi
func ValidateInput(engine *FlowEngine, minLength int) error {
	fmt.Printf("Validating input with min length: %d\n", minLength)
	if name, exists := engine.State["name"]; exists {
		if len(name.(string)) < minLength {
			return fmt.Errorf("name too short, minimum %d characters", minLength)
		}
	}
	return nil
}

func GenerateDocument(engine *FlowEngine, templateName string, data map[string]any) (string, error) {
	fmt.Printf("Generating document with template: %s and data: %v\n", templateName, data)
	// Simpan hasil ke state
	docID := "doc_12345"
	engine.State["document_id"] = docID
	return docID, nil
}

func main() {
	// 1. Inisialisasi Flow Engine
	engine := NewFlowEngine()

	// 2. Registrasi function-function yang tersedia
	engine.RegisterFunction("validate_input", ValidateInput)
	engine.RegisterFunction("generate_document", GenerateDocument)

	// 3. Load flow definition dari JSON
	err := engine.LoadFlowFromFile("flow_config.json")
	if err != nil {
		panic(err)
	}

	// 4. Set initial state
	engine.State["name"] = "John Doe"
	engine.State["age"] = 30
	engine.State["address"] = "123 Main St"

	// 5. Eksekusi flow
	err = engine.Execute()
	if err != nil {
		fmt.Println("Flow execution failed:", err)
	} else {
		fmt.Println("Flow executed successfully")
		fmt.Println("Final state:", engine.State)
	}
}
