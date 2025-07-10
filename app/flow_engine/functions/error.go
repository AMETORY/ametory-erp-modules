package functions

import (
	"fmt"
	"time"

	. "github.com/AMETORY/ametory-erp-modules/app/flow_engine"
)

func LogError(engine *FlowEngine, errorType string, errorDetails interface{}) error {
	logEntry := map[string]interface{}{
		"timestamp":    time.Now().Format(time.RFC3339),
		"error_type":   errorType,
		"details":      errorDetails,
		"current_step": engine.State["last_executed_step"],
	}

	// Simpan log di state
	if _, exists := engine.State["error_logs"]; !exists {
		engine.State["error_logs"] = []map[string]interface{}{}
	}

	logs := engine.State["error_logs"].([]map[string]interface{})
	engine.State["error_logs"] = append(logs, logEntry)

	fmt.Printf("\nLogged error: %s - %v\n", errorType, errorDetails)
	return nil
}
