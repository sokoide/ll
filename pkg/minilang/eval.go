package minilang

import (
	"fmt"
	"io"
	"strconv"
)

// evaluator is the tree-walk baseline: execution recursively visits AST nodes.
type evaluator struct {
	scopes *ScopeStack
	output io.Writer
}

func newEvaluator(output io.Writer) *evaluator {
	return &evaluator{scopes: NewScopeStack(), output: output}
}

func (e *evaluator) evaluateStmts(stmts []statement) (int, error) {
	for _, stmt := range stmts {
		if _, err := e.evaluateStmt(stmt); err != nil {
			return 0, err
		}
	}
	return 0, nil
}

func (e *evaluator) evaluateStmt(stmt statement) (int, error) {
	switch s := stmt.(type) {
	case *assignStatement:
		value, err := e.evaluateExpr(s.Expr)
		if err != nil {
			return 0, err
		}
		e.scopes.assign(s.Name, value)
	case *putStatement:
		for _, expr := range s.Exprs {
			if err := e.printExpr(expr); err != nil {
				return 0, err
			}
		}
		_, err := fmt.Fprintln(e.output)
		return 0, err
	case *whileStatement:
		e.scopes.PushScope()
		defer e.scopes.PopScope()
		for {
			value, err := e.evaluateExpr(s.Cond)
			if err != nil {
				return 0, err
			}
			if value <= 0 {
				break
			}
			if _, err := e.evaluateStmts(s.Body); err != nil {
				return 0, err
			}
		}
	case *emptyStatement:
		// An empty statement has no effect.
	default:
		return 0, fmt.Errorf("unsupported statement %T", stmt)
	}
	return 0, nil
}

func (e *evaluator) printExpr(expr expression) error {
	// Direct literals retain their spelling, including leading zeros.
	switch x := expr.(type) {
	case *numberExpression:
		_, err := io.WriteString(e.output, x.Lit)
		return err
	case *stringExpression:
		_, err := io.WriteString(e.output, x.Lit)
		return err
	default:
		value, err := e.evaluateExpr(expr)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(e.output, value)
		return err
	}
}

func (e *evaluator) evaluateExpr(expr expression) (int, error) {
	switch x := expr.(type) {
	case *numberExpression:
		return strconv.Atoi(x.Lit)
	case *variableExpression:
		return e.scopes.load(x.Lit), nil
	case *parenExpression:
		return e.evaluateExpr(x.SubExpr)
	case *binOpExpression:
		return e.evaluateBinary(x.LHS, x.Operator, x.RHS)
	case *condExpression:
		return e.evaluateBinary(x.LHS, x.Operator, x.RHS)
	default:
		return 0, fmt.Errorf("unsupported numeric expression %T", expr)
	}
}

func (e *evaluator) evaluateBinary(lhs expression, op int, rhs expression) (int, error) {
	left, err := e.evaluateExpr(lhs)
	if err != nil {
		return 0, err
	}
	right, err := e.evaluateExpr(rhs)
	if err != nil {
		return 0, err
	}
	return applyBinary(op, left, right)
}
