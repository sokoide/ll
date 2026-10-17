// Package minilang parses and executes the language defined in grammar.ebnf.
package minilang

import (
	"fmt"
	"io"
)

// Program is a parsed program. Its AST is immutable during execution.
type Program struct {
	statements []statement
}

// Parse implements program = stmts, EOF, returning an error before execution
// if any part of the source is invalid.
func Parse(source string) (*Program, error) {
	s := new(scanner)
	s.Init(source)
	statements, err := parse(s)
	if err != nil {
		return nil, err
	}
	return &Program{statements: statements}, nil
}

// Execute runs the program with fresh variable scopes and writes put output to
// output. Separate executions do not share variables.
func (p *Program) Execute(output io.Writer) error {
	if output == nil {
		return fmt.Errorf("output writer must not be nil")
	}
	_, err := newEvaluator(output).evaluateStmts(p.statements)
	return err
}
