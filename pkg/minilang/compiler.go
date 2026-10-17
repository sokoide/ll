package minilang

import "fmt"

type opcode int

const (
	opPush opcode = iota
	opLoad
	opStore
	opBinary
	opPrintInt
	opPrintText
	opNewline
	opEnterScope
	opLeaveScope
	opJump
	opJumpIfFalse
)

// instruction is an in-memory bytecode instruction, not a binary file format.
// text holds a literal or variable name; arg holds an operator or jump address.
type instruction struct {
	op   opcode
	text string
	arg  int
}

// Bytecode is immutable compiled code. Execution state belongs to each VM run.
type Bytecode struct{ code []instruction }

// Compile lowers an AST to stack instructions; parsing never executes code.
func (p *Program) Compile() (*Bytecode, error) {
	c := &compiler{}
	if err := c.compileStmts(p.statements); err != nil {
		return nil, err
	}
	return &Bytecode{code: c.code}, nil
}

type compiler struct{ code []instruction }

// emit returns the address so a forward jump can be patched later.
func (c *compiler) emit(op opcode, text string, arg int) int {
	address := len(c.code)
	c.code = append(c.code, instruction{op: op, text: text, arg: arg})
	return address
}

// compileStmts implements stmts = { stmt } in source order.
func (c *compiler) compileStmts(stmts []statement) error {
	for _, stmt := range stmts {
		if err := c.compileStmt(stmt); err != nil {
			return err
		}
	}
	return nil
}

// compileStmt lowers assignment, put_stmt, while_stmt, and empty statements.
// Every complete statement has net zero effect on the operand stack.
func (c *compiler) compileStmt(stmt statement) error {
	switch s := stmt.(type) {
	case *assignStatement:
		if err := c.compileExpr(s.Expr); err != nil {
			return err
		}
		c.emit(opStore, s.Name, 0)
	case *putStatement:
		for _, expr := range s.Exprs {
			switch x := expr.(type) {
			case *stringExpression:
				c.emit(opPrintText, x.Lit, 0)
			case *numberExpression:
				c.emit(opPrintText, x.Lit, 0)
			default:
				if err := c.compileExpr(expr); err != nil {
					return err
				}
				c.emit(opPrintInt, "", 0)
			}
		}
		c.emit(opNewline, "", 0)
	case *whileStatement:
		// Enter once for the whole loop, not once per iteration.
		c.emit(opEnterScope, "", 0)
		start := len(c.code)
		if err := c.compileExpr(s.Cond); err != nil {
			return err
		}
		exit := c.emit(opJumpIfFalse, "", 0)
		if err := c.compileStmts(s.Body); err != nil {
			return err
		}
		c.emit(opJump, "", start)
		c.code[exit].arg = len(c.code) // Forward target is now known.
		c.emit(opLeaveScope, "", 0)
	case *emptyStatement:
		// No instruction is needed for ';'.
	default:
		return fmt.Errorf("unsupported statement %T", stmt)
	}
	return nil
}

// compileExpr lowers expr, term, primary, and cond in postorder.
// Each expression leaves exactly one integer on the operand stack.
func (c *compiler) compileExpr(expr expression) error {
	switch x := expr.(type) {
	case *numberExpression:
		c.emit(opPush, x.Lit, 0)
	case *variableExpression:
		c.emit(opLoad, x.Lit, 0)
	case *parenExpression:
		return c.compileExpr(x.SubExpr)
	case *binOpExpression:
		return c.compileBinary(x.LHS, x.Operator, x.RHS)
	case *condExpression:
		return c.compileBinary(x.LHS, x.Operator, x.RHS)
	default:
		return fmt.Errorf("unsupported numeric expression %T", expr)
	}
	return nil
}

// compileBinary emits lhs then rhs; BINARY consumes rhs first, then lhs.
func (c *compiler) compileBinary(lhs expression, op int, rhs expression) error {
	if err := c.compileExpr(lhs); err != nil {
		return err
	}
	if err := c.compileExpr(rhs); err != nil {
		return err
	}
	c.emit(opBinary, "", op)
	return nil
}
