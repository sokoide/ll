package minilang

import "fmt"

// applyBinary defines integer semantics shared by the tree evaluator and VM.
// It knows nothing about parsing, AST traversal, or instruction execution.
func applyBinary(op, lhs, rhs int) (int, error) {
	var result bool
	switch op {
	case '+':
		return lhs + rhs, nil
	case '-':
		return lhs - rhs, nil
	case '*':
		return lhs * rhs, nil
	case '/':
		if rhs == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return lhs / rhs, nil
	case '%':
		if rhs == 0 {
			return 0, fmt.Errorf("modulo by zero")
		}
		return lhs % rhs, nil
	case EQOP:
		result = lhs == rhs
	case NEOP:
		result = lhs != rhs
	case GEOP:
		result = lhs >= rhs
	case GTOP:
		result = lhs > rhs
	case LEOP:
		result = lhs <= rhs
	case LTOP:
		result = lhs < rhs
	default:
		return 0, fmt.Errorf("unknown operator %d", op)
	}
	if result {
		return 1, nil
	}
	return 0, nil
}
