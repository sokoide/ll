package minilang

import (
	"fmt"
	"io"
	"reflect"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func (e *evaluator) evaluateStmts(stmts []expression) (int, error) {
	var ret int
	var err error

	for line, stmt := range stmts {
		ret, err = e.evaluateStmt(stmt)
		if err != nil {
			log.Errorf("[stmt %d] evaluation error: %v", line, err)
			return 0, err
		}
	}
	return ret, err
}

func (e *evaluator) evaluateStmt(stmt expression) (int, error) {
	var ret int
	var err error

	switch s := stmt.(type) {
	case *exprStatement:
		ret, err = e.evaluateExpr(s.Expr)
		if err == nil {
			log.Debugf("expr result: %d\n", ret)
			return 0, nil
		}
		return 0, err
	case *assignStatement:
		ret, err = e.evaluateExpr(s.Expr)
		if err == nil {
			if _, scope := e.scopes.Get(s.Name); scope != nil {
				scope.Set(s.Name, ret)
			} else {
				log.Debugf("%s doesn't exist in any scope. Making it in the current scope", s.Name)
				e.scopes.Set(s.Name, ret)
			}
			log.Debugf("%s = %d\n", s.Name, ret)
			return 0, nil
		}
		return 0, err
	case *putStatement:
		for _, expr := range s.Exprs {
			if err := e.printExpr(expr); err != nil {
				return 0, err
			}
		}
		_, err := fmt.Fprintln(e.output)
		return 0, err
	case *whileStatement:
		log.Debug("whileStatement - scope pushed")
		e.scopes.PushScope()
		for {
			ret, err = e.evaluateExpr(s.Cond)
			if err != nil {
				log.Debug("whileStatement -scope poped")
				e.scopes.PopScope()
				return 0, err
			}
			if ret > 0 {
				_, err = e.evaluateStmts(s.Body)
				if err != nil {
					log.Debug("whileStatement -scope poped")
					e.scopes.PopScope()
					return 0, err
				}
			} else {
				break
			}
		}
		log.Debug("whileStatement -scope poped")
		e.scopes.PopScope()
		return 0, err
	case *emptyStatement:
		log.Debug("emptyStatement")
		return 0, nil
	default:
		return 0, fmt.Errorf("%s not supported yet", reflect.TypeOf(stmt))
	}
}

func (e *evaluator) printExpr(expr expression) error {
	log.Debugf("printExpr %+v\n", expr)

	var ret int
	var err error
	var scope *Scope

	switch x := expr.(type) {
	case *numberExpression:
		_, err = fmt.Fprintf(e.output, "%s", x.Lit)
	case *variableExpression:
		if ret, scope = e.scopes.Get(x.Lit); scope != nil {
			_, err = fmt.Fprintf(e.output, "%d", ret)
		} else {
			// Same semantics as evaluateExpr: reading an undefined
			// variable yields 0 (and auto-creates it in the current
			// scope) instead of failing.
			e.scopes.Set(x.Lit, 0)
			_, err = fmt.Fprintf(e.output, "%d", 0)
		}
	case *stringExpression:
		_, err = fmt.Fprintf(e.output, "%s", x.Lit)
	default:
		ret, err = e.evaluateExpr(expr)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(e.output, "%d", ret)
	}
	return err
}

func (e *evaluator) evaluateExpr(expr expression) (int, error) {
	log.Debugf("evaluateExpr %+v\n", expr)
	switch x := expr.(type) {

	case *numberExpression:
		v, err := strconv.Atoi(x.Lit)
		if err != nil {
			log.Warnf("err: %v", err)
			return 0, err
		}
		return v, nil
	case *parenExpression:
		v, err := e.evaluateExpr(x.SubExpr)
		if err != nil {
			log.Warnf("err: %v", err)
			return 0, err
		}
		return v, nil
	case *binOpExpression:
		lhsV, err := e.evaluateExpr(x.LHS)
		if err != nil {
			log.Warnf("err: %v", err)
			return 0, err
		}
		rhsV, err := e.evaluateExpr(x.RHS)
		if err != nil {
			log.Warnf("err: %v", err)
			return 0, err
		}
		switch x.Operator {
		case '+':
			return lhsV + rhsV, nil
		case '-':
			return lhsV - rhsV, nil
		case '*':
			return lhsV * rhsV, nil
		case '/':
			if rhsV == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return lhsV / rhsV, nil
		case '%':
			if rhsV == 0 {
				return 0, fmt.Errorf("modulo by zero")
			}
			return lhsV % rhsV, nil
		default:
			return 0, fmt.Errorf("unknown operator: %v", x.Operator)
		}
	case *variableExpression:
		var v int
		var scope *Scope
		if v, scope = e.scopes.Get(x.Lit); scope != nil {
			return v, nil
		}
		// Reading an undefined variable yields 0 (and auto-creates it in
		// the current scope so assignments like `b=b+1;` inside a while
		// body still work).
		e.scopes.Set(x.Lit, 0)
		log.Warnf("variable %s not found, 0 assigned", x.Lit)
		return 0, nil
	case *stringExpression:
		return 0, fmt.Errorf("string %q not supported in arithmetic", x.Lit)
	case *condExpression:
		lhsV, err := e.evaluateExpr(x.LHS)
		if err != nil {
			log.Warnf("err: %v", err)
			return 0, err
		}
		rhsV, err := e.evaluateExpr(x.RHS)
		if err != nil {
			log.Warnf("err: %v", err)
			return 0, err
		}
		switch x.Operator {
		case EQOP:
			if lhsV == rhsV {
				return 1, nil
			}
			return 0, nil
		case NEOP:
			if lhsV != rhsV {
				return 1, nil
			}
			return 0, nil
		case GEOP:
			if lhsV >= rhsV {
				return 1, nil
			}
			return 0, nil
		case GTOP:
			if lhsV > rhsV {
				return 1, nil
			}
			return 0, nil
		case LEOP:
			if lhsV <= rhsV {
				return 1, nil
			}
			return 0, nil
		case LTOP:
			if lhsV < rhsV {
				return 1, nil
			}
			return 0, nil
		default:
			return 0, fmt.Errorf("invalid Operator %v", x.Operator)
		}
	default:
		return 0, fmt.Errorf("unknown expression type %s for %+v", reflect.TypeOf(x), x)
	}
}

// evaluator owns the mutable state of one program execution.
type evaluator struct {
	scopes *ScopeStack
	output io.Writer
}

func newEvaluator(output io.Writer) *evaluator {
	return &evaluator{scopes: NewScopeStack(), output: output}
}
