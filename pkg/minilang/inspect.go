package minilang

import (
	"fmt"
	"io"
	"strings"
)

// WriteTokens shows the lexer's terminal stream, including EOF, without parsing.
func WriteTokens(source string, output io.Writer) error {
	s := new(scanner)
	s.Init(source)
	for {
		kind, lit, pos, _ := s.Scan()
		if _, err := fmt.Fprintf(output, "%d:%d %-8s %q\n", pos.Line, pos.Column, tokenName(kind), lit); err != nil {
			return err
		}
		if kind == ILLEGAL {
			return fmt.Errorf("[Line:%d Column:%d] illegal token %q", pos.Line, pos.Column, lit)
		}
		if kind == EOF {
			return nil
		}
	}
}

// WriteAST shows statements and expressions before evaluation or compilation.
func (p *Program) WriteAST(output io.Writer) error {
	var text strings.Builder
	for _, stmt := range p.statements {
		formatStmt(&text, stmt, 0)
	}
	_, err := io.WriteString(output, text.String())
	return err
}

func formatStmt(out *strings.Builder, stmt statement, depth int) {
	indent := strings.Repeat("  ", depth)
	switch s := stmt.(type) {
	case *assignStatement:
		fmt.Fprintf(out, "%sAssign %s\n", indent, s.Name)
		formatExpr(out, s.Expr, depth+1)
	case *putStatement:
		fmt.Fprintf(out, "%sPut\n", indent)
		for _, expr := range s.Exprs {
			formatExpr(out, expr, depth+1)
		}
	case *whileStatement:
		fmt.Fprintf(out, "%sWhile\n%s  Condition\n", indent, indent)
		formatExpr(out, s.Cond, depth+2)
		fmt.Fprintf(out, "%s  Body\n", indent)
		for _, stmt := range s.Body {
			formatStmt(out, stmt, depth+2)
		}
	case *emptyStatement:
		fmt.Fprintf(out, "%sEmpty\n", indent)
	}
}

func formatExpr(out *strings.Builder, expr expression, depth int) {
	indent := strings.Repeat("  ", depth)
	switch x := expr.(type) {
	case *numberExpression:
		fmt.Fprintf(out, "%sNumber %s\n", indent, x.Lit)
	case *variableExpression:
		fmt.Fprintf(out, "%sVariable %s\n", indent, x.Lit)
	case *stringExpression:
		fmt.Fprintf(out, "%sString %q\n", indent, x.Lit)
	case *parenExpression:
		fmt.Fprintf(out, "%sParen\n", indent)
		formatExpr(out, x.SubExpr, depth+1)
	case *binOpExpression:
		fmt.Fprintf(out, "%sBinary %s\n", indent, tokenName(x.Operator))
		formatExpr(out, x.LHS, depth+1)
		formatExpr(out, x.RHS, depth+1)
	case *condExpression:
		fmt.Fprintf(out, "%sCompare %s\n", indent, tokenName(x.Operator))
		formatExpr(out, x.LHS, depth+1)
		formatExpr(out, x.RHS, depth+1)
	}
}
