package flow_engine

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"
)

// 1. Definisi struktur Flow Engine
type FlowEngine struct {
	State     map[string]any
	Functions map[string]any
	Templates map[string]string
	Steps     []FlowStep
}

type FlowStep struct {
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Type          string         `json:"type"` // "function", "api_call", "conditional"
	Function      string         `json:"function,omitempty"`
	Params        map[string]any `json:"params"`
	NextOnSuccess string         `json:"next_on_success,omitempty"`
	NextOnError   string         `json:"next_on_error,omitempty"`
}

// 2. Inisialisasi Flow Engine
func NewFlowEngine() *FlowEngine {
	return &FlowEngine{
		State:     make(map[string]any),
		Functions: make(map[string]any),
		Templates: make(map[string]string),
	}
}

// 3. Registrasi Function
func (e *FlowEngine) RegisterFunction(name string, fn any) {
	e.Functions[name] = fn
}

// 4. Load Flow dari JSON
func (e *FlowEngine) LoadFlowFromFile(filename string) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var steps []FlowStep
	err = json.Unmarshal(file, &steps)
	if err != nil {
		return err
	}

	e.Steps = steps
	return nil
}

// 5. Eksekusi Flow
func (e *FlowEngine) Execute() error {
	for _, step := range e.Steps {
		// fmt.Printf("Executing step: %s (%s)\n", step.Name, step.Type)

		var err error
		switch step.Type {
		case "function":
			err = e.ExecuteFunctionStep(step)
		case "api_call":
			err = e.ExecuteAPICallStep(step)
		case "conditional":
			err = e.ExecuteConditionalStep(step)
		default:
			err = fmt.Errorf("unknown step type: %s", step.Type)
		}

		if err != nil {
			fmt.Printf("Error executing step %s: %v\n", step.Name, err)
			if step.NextOnError != "" {
				if nextStep := e.FindStepByName(step.NextOnError); nextStep != nil {
					return e.ExecuteSingleStep(*nextStep)
				}
			}
			return err
		}

		if step.NextOnSuccess != "" {
			if nextStep := e.FindStepByName(step.NextOnSuccess); nextStep != nil {
				return e.ExecuteSingleStep(*nextStep)
			}
		}
	}
	return nil
}

// 6. Eksekusi Function Step dengan Reflection
func (e *FlowEngine) ExecuteFunctionStep(step FlowStep) error {
	fn, exists := e.Functions[step.Function]
	if !exists {
		return fmt.Errorf("function %s not registered", step.Function)
	}

	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() != reflect.Func {
		return fmt.Errorf("%s is not a function", step.Function)
	}

	// Proses parameter
	params, err := e.prepareFunctionParams(fnValue, step.Params)
	if err != nil {
		return fmt.Errorf("error preparing params: %v", err)
	}

	// Panggil function
	results := fnValue.Call(params)

	// Proses return value
	return e.processFunctionResults(step, results)
}

// 7. Persiapan Parameter Function
func (e *FlowEngine) prepareFunctionParams(fnValue reflect.Value, stepParams map[string]any) ([]reflect.Value, error) {
	fnType := fnValue.Type()
	numIn := fnType.NumIn()
	params := make([]reflect.Value, numIn)
	// fmt.Println("numIn", numIn)

	for i := range numIn {
		paramType := fnType.In(i)
		// Handle context injection (FlowEngine pointer)
		if paramType.String() == "*flow_engine.FlowEngine" {
			params[i] = reflect.ValueOf(e)
			continue
		}

		// Handle parameter dari step.Params
		paramName := fmt.Sprintf("arg%d", i-1)
		// fmt.Println("stepParams", paramName, i)
		paramValue, exists := stepParams[paramName]
		if !exists {
			return nil, fmt.Errorf("missing parameter %s", paramName)
		}

		// Convert parameter sesuai type yang dibutuhkan
		val, err := e.convertParam(paramValue, paramType)
		if err != nil {
			return nil, fmt.Errorf("error converting param %s: %v", paramName, err)
		}
		params[i] = val
	}

	return params, nil
}

// 8. Konversi Parameter
func (e *FlowEngine) convertParam(value any, targetType reflect.Type) (reflect.Value, error) {
	// Handle state variable (format: ${state.key})
	if strVal, ok := value.(string); ok && strings.HasPrefix(strVal, "${") && strings.HasSuffix(strVal, "}") {
		key := strings.TrimSuffix(strings.TrimPrefix(strVal, "${"), "}")
		if stateVal, exists := e.State[key]; exists {
			value = stateVal
		} else {
			return reflect.Value{}, fmt.Errorf("state variable %s not found", key)
		}
	}

	val := reflect.ValueOf(value)
	if !val.Type().ConvertibleTo(targetType) {
		return reflect.Value{}, fmt.Errorf("cannot convert %v to %v", val.Type(), targetType)
	}

	return val.Convert(targetType), nil
}

// 9. Proses Hasil Function
func (e *FlowEngine) processFunctionResults(step FlowStep, results []reflect.Value) error {
	// Simpan hasil ke state jika ada
	if storeKey, ok := step.Params["_store_result"].(string); ok && len(results) > 0 {
		e.State[storeKey] = results[0].Interface()
	}

	// Handle error return
	if len(results) > 0 {
		if err, ok := results[len(results)-1].Interface().(error); ok && err != nil {
			return err
		}
	}

	return nil
}

func (e *FlowEngine) FindStepByName(name string) *FlowStep {
	for _, step := range e.Steps {
		if step.Name == name {
			return &step
		}
	}
	return nil
}

func (e *FlowEngine) ExecuteSingleStep(step FlowStep) error {
	// 1. Log eksekusi step
	// fmt.Printf("[%s] Executing step: %s (%s)\n", time.Now().Format(time.RFC3339), step.Name, step.Type)

	// 2. Eksekusi berdasarkan tipe step
	var err error
	switch step.Type {
	case "function":
		err = e.ExecuteFunctionStep(step)
	case "api_call":
		err = e.ExecuteAPICallStep(step)
	case "conditional":
		err = e.ExecuteConditionalStep(step)
	case "delay":
		err = e.executeDelayStep(step)
	case "parallel":
		err = e.executeParallelStep(step)
	case "wait_input":
		// Tidak perlu melakukan apa-apa, hanya menunggu input
		return nil
	default:
		err = fmt.Errorf("unknown step type: %s", step.Type)
	}

	// 3. Handle error jika terjadi
	if err != nil {
		// Log error
		fmt.Printf("[%s] Error executing step %s: %v\n", time.Now().Format(time.RFC3339), step.Name, err)

		// Simpan error ke state jika diperlukan
		e.State["last_error"] = map[string]interface{}{
			"step":  step.Name,
			"error": err.Error(),
			"time":  time.Now().Format(time.RFC3339),
		}

		return err
	}

	// 4. Simpan eksekusi terakhir ke state
	e.State["last_executed_step"] = step.Name
	e.State["last_execution_time"] = time.Now().Format(time.RFC3339)

	if step.NextOnSuccess != "" {
		if nextStep := e.FindStepByName(step.NextOnSuccess); nextStep != nil {
			return e.ExecuteSingleStep(*nextStep)
		}
	}

	return nil
}

// Implementasi untuk tipe step tambahan
func (e *FlowEngine) executeDelayStep(step FlowStep) error {
	// 1. Validasi parameter
	if step.Params["duration"] == nil {
		return fmt.Errorf("missing duration parameter")
	}

	// 2. Parse durasi
	durationStr, ok := step.Params["duration"].(string)
	if !ok {
		return fmt.Errorf("duration should be a string (e.g., '5s', '1m')")
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return fmt.Errorf("invalid duration format: %v", err)
	}

	// 3. Jalankan delay
	fmt.Printf("Delaying for %v...\n", duration)
	time.Sleep(duration)
	return nil
}

func (e *FlowEngine) executeParallelStep(step FlowStep) error {
	// 1. Validasi parameter
	if step.Params["steps"] == nil {
		return fmt.Errorf("missing steps parameter for parallel execution")
	}

	stepsList, ok := step.Params["steps"].([]interface{})
	if !ok {
		return fmt.Errorf("steps should be an array of step names")
	}

	// 2. Konversi ke slice string
	var stepNames []string
	for _, s := range stepsList {
		if name, ok := s.(string); ok {
			stepNames = append(stepNames, name)
		} else {
			return fmt.Errorf("invalid step name in parallel steps list")
		}
	}

	// 3. Temukan step-step yang akan dijalankan
	var stepsToExecute []FlowStep
	for _, name := range stepNames {
		foundStep := e.FindStepByName(name)
		if foundStep == nil {
			return fmt.Errorf("step %s not found for parallel execution", name)
		}
		stepsToExecute = append(stepsToExecute, *foundStep)
	}

	// 4. Jalankan secara parallel
	var wg sync.WaitGroup
	errChan := make(chan error, len(stepsToExecute))

	for _, step := range stepsToExecute {
		wg.Add(1)
		go func(s FlowStep) {
			defer wg.Done()
			if err := e.ExecuteSingleStep(s); err != nil {
				errChan <- fmt.Errorf("error in parallel step %s: %v", s.Name, err)
			}
		}(step)
	}

	// 5. Tunggu semua goroutine selesai
	wg.Wait()
	close(errChan)

	// 6. Kumpulkan semua error
	var errors []string
	for err := range errChan {
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		return fmt.Errorf("parallel execution errors:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}
