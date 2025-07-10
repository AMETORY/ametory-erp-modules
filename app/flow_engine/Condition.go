package flow_engine

import (
	"fmt"
	"strconv"
	"strings"
)

func (e *FlowEngine) ExecuteConditionalStep(step FlowStep) error {
	// 1. Validate required params
	if step.Params["condition"] == nil {
		return fmt.Errorf("missing condition parameter")
	}

	condition, ok := step.Params["condition"].(string)
	if !ok {
		return fmt.Errorf("condition should be a string")
	}

	// 2. Parse condition
	expr, err := parseCondition(condition)
	if err != nil {
		return fmt.Errorf("error parsing condition: %v", err)
	}

	// 3. Evaluate condition
	result, err := expr.Evaluate(e.State)
	if err != nil {
		return fmt.Errorf("error evaluating condition: %v", err)
	}

	// 4. Determine next step
	var nextStep string
	if result {
		nextStep, ok = step.Params["true_step"].(string)
		if !ok {
			return fmt.Errorf("true_step not specified or invalid")
		}
	} else {
		nextStep, ok = step.Params["false_step"].(string)
		if !ok {
			return fmt.Errorf("false_step not specified or invalid")
		}
	}

	// 5. Find and execute next step
	targetStep := e.FindStepByName(nextStep)
	if targetStep == nil {
		return fmt.Errorf("target step %s not found", nextStep)
	}

	return e.ExecuteSingleStep(*targetStep)
}

// Condition expression parsing and evaluation
type ConditionExpr interface {
	Evaluate(state map[string]interface{}) (bool, error)
}

type AndExpr struct {
	Left, Right ConditionExpr
}

func (e AndExpr) Evaluate(state map[string]interface{}) (bool, error) {
	leftVal, err := e.Left.Evaluate(state)
	if err != nil {
		return false, err
	}
	if !leftVal {
		return false, nil
	}
	return e.Right.Evaluate(state)
}

type OrExpr struct {
	Left, Right ConditionExpr
}

func (e OrExpr) Evaluate(state map[string]interface{}) (bool, error) {
	leftVal, err := e.Left.Evaluate(state)
	if err != nil {
		return false, err
	}
	if leftVal {
		return true, nil
	}
	return e.Right.Evaluate(state)
}

type ComparisonExpr struct {
	Left, Right string
	Operator    string
}

func (e ComparisonExpr) Evaluate(state map[string]interface{}) (bool, error) {
	leftVal, err := getValueFromState(e.Left, state)
	if err != nil {
		return false, err
	}

	rightVal, err := getValueFromState(e.Right, state)
	if err != nil {
		return false, err
	}

	switch e.Operator {
	case "==":
		return compareValues(leftVal, rightVal) == 0, nil
	case "!=":
		return compareValues(leftVal, rightVal) != 0, nil
	case ">":
		return compareValues(leftVal, rightVal) > 0, nil
	case ">=":
		return compareValues(leftVal, rightVal) >= 0, nil
	case "<":
		return compareValues(leftVal, rightVal) < 0, nil
	case "<=":
		return compareValues(leftVal, rightVal) <= 0, nil
	default:
		return false, fmt.Errorf("unknown operator: %s", e.Operator)
	}
}

func parseCondition(condStr string) (ConditionExpr, error) {
	// Simple parser for demonstration
	// In real implementation, consider using a proper parser library

	// Check for AND/OR conditions
	if strings.Contains(condStr, "&&") {
		parts := strings.SplitN(condStr, "&&", 2)
		left, err := parseCondition(strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, err
		}
		right, err := parseCondition(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, err
		}
		return AndExpr{Left: left, Right: right}, nil
	}

	if strings.Contains(condStr, "||") {
		parts := strings.SplitN(condStr, "||", 2)
		left, err := parseCondition(strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, err
		}
		right, err := parseCondition(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, err
		}
		return OrExpr{Left: left, Right: right}, nil
	}

	// Handle comparison operators
	for _, op := range []string{"==", "!=", ">=", "<=", ">", "<"} {
		if strings.Contains(condStr, op) {
			parts := strings.SplitN(condStr, op, 2)
			if len(parts) == 2 {
				return ComparisonExpr{
					Left:     strings.TrimSpace(parts[0]),
					Operator: op,
					Right:    strings.TrimSpace(parts[1]),
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("could not parse condition: %s", condStr)
}

func getValueFromState(key string, state map[string]interface{}) (interface{}, error) {
	// Handle literal values
	if strings.HasPrefix(key, "'") && strings.HasSuffix(key, "'") {
		return strings.Trim(key, "'"), nil
	}
	if strings.HasPrefix(key, `"`) && strings.HasSuffix(key, `"`) {
		return strings.Trim(key, `"`), nil
	}
	if _, err := strconv.Atoi(key); err == nil {
		return strconv.Atoi(key)
	}
	if _, err := strconv.ParseFloat(key, 64); err == nil {
		return strconv.ParseFloat(key, 64)
	}

	// Handle state variables
	val, exists := state[key]
	if !exists {
		return nil, fmt.Errorf("variable %s not found in state", key)
	}
	return val, nil
}

func compareValues(a, b interface{}) int {
	// Simple comparison for demonstration
	// In real implementation, handle more types and cases
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return strings.Compare(aStr, bStr)
}
