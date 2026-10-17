package minilang

import "fmt"

// parser uses one lookahead token, with no backtracking.
type parser struct {
	s       *scanner
	current token
}

// parse implements program = stmts, EOF.
func parse(s *scanner) ([]expression, error) {
	p := &parser{s: s}
	p.advance()
	stmts, err := p.parseStmts()
	if err != nil {
		return nil, err
	}
	if err := p.expect(EOF); err != nil {
		return nil, err
	}
	return stmts, nil
}

// advance is the token-reading helper for all grammar productions.
func (p *parser) advance() {
	tok, lit, pos, tt := p.s.Scan()
	p.current = token{tok: tok, lit: lit, pos: pos, tt: tt}
}

// unexpected reports the lookahead that cannot match a grammar production.
func (p *parser) unexpected(want string) error {
	return fmt.Errorf("[Line:%d Column:%d] expected %s, got %q (token %d)", p.current.pos.Line, p.current.pos.Column, want, p.current.lit, p.current.tok)
}

// expect consumes a required terminal in the calling grammar production.
func (p *parser) expect(tok int) error {
	if p.current.tok != tok {
		return p.unexpected(fmt.Sprintf("token %d", tok))
	}
	p.advance()
	return nil
}

// parseStmts implements stmts = { stmt }; its FOLLOW set is '}' or EOF.
func (p *parser) parseStmts() ([]expression, error) {
	var result []expression
	for p.current.tok != EOF && p.current.tok != '}' {
		stmt, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		result = append(result, stmt)
	}
	return result, nil
}

// parseStmt implements stmt = assignment | while_stmt | put_stmt | ';'.
func (p *parser) parseStmt() (expression, error) {
	switch p.current.tok {
	case IDENT:
		return p.parseAssignment()
	case WHILE:
		return p.parseWhile()
	case PUT:
		return p.parsePut()
	case ';':
		p.advance()
		return &emptyStatement{}, nil
	default:
		return nil, p.unexpected("assignment, while, put or ';'")
	}
}

// parseAssignment implements assignment = IDENT, '=', expr, ';'.
func (p *parser) parseAssignment() (expression, error) {
	name := p.current.lit
	p.advance()
	if err := p.expect('='); err != nil {
		return nil, err
	}
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if err := p.expect(';'); err != nil {
		return nil, err
	}
	return &assignStatement{Name: name, Expr: expr}, nil
}

// parseWhile implements while_stmt = 'while', '(', cond, ')', '{', stmts, '}'.
func (p *parser) parseWhile() (expression, error) {
	p.advance()
	if err := p.expect('('); err != nil {
		return nil, err
	}
	cond, err := p.parseCond()
	if err != nil {
		return nil, err
	}
	if err := p.expect(')'); err != nil {
		return nil, err
	}
	if err := p.expect('{'); err != nil {
		return nil, err
	}
	body, err := p.parseStmts()
	if err != nil {
		return nil, err
	}
	if err := p.expect('}'); err != nil {
		return nil, err
	}
	return &whileStatement{Cond: cond, Body: body}, nil
}

// parsePut implements put_stmt = 'put', '(', put_list, ')', ';'.
func (p *parser) parsePut() (expression, error) {
	p.advance()
	if err := p.expect('('); err != nil {
		return nil, err
	}
	exprs, err := p.parsePutList()
	if err != nil {
		return nil, err
	}
	if err := p.expect(')'); err != nil {
		return nil, err
	}
	if err := p.expect(';'); err != nil {
		return nil, err
	}
	return &putStatement{Exprs: exprs}, nil
}

// parsePutList implements put_list = expr, { ',', expr }.
func (p *parser) parsePutList() ([]expression, error) {
	first, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	result := []expression{first}
	for p.current.tok == ',' {
		p.advance()
		next, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		result = append(result, next)
	}
	return result, nil
}

// parseCond implements cond = expr, [ compare_op, expr ].
// compare_op = '==' | '!=' | '>=' | '>' | '<=' | '<'.
func (p *parser) parseCond() (expression, error) {
	lhs, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	switch p.current.tok {
	case EQOP, NEOP, GEOP, GTOP, LEOP, LTOP:
		op := p.current.tok
		p.advance()
		rhs, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		return &condExpression{LHS: lhs, Operator: op, RHS: rhs}, nil
	}
	return lhs, nil
}

// parseExpr implements expr = term, { ('+' | '-'), term }.
// The loop builds a left-associated AST without left recursion.
func (p *parser) parseExpr() (expression, error) {
	lhs, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	for p.current.tok == '+' || p.current.tok == '-' {
		op := p.current.tok
		p.advance()
		rhs, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		lhs = &binOpExpression{LHS: lhs, Operator: op, RHS: rhs}
	}
	return lhs, nil
}

// parseTerm implements term = primary, { ('*' | '/' | '%'), primary }.
func (p *parser) parseTerm() (expression, error) {
	lhs, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	for p.current.tok == '*' || p.current.tok == '/' || p.current.tok == '%' {
		op := p.current.tok
		p.advance()
		rhs, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		lhs = &binOpExpression{LHS: lhs, Operator: op, RHS: rhs}
	}
	return lhs, nil
}

// parsePrimary implements primary = NUMBER | IDENT | STRING | '(', expr, ')'.
func (p *parser) parsePrimary() (expression, error) {
	t := p.current
	switch t.tok {
	case NUMBER_LITERAL:
		p.advance()
		return &numberExpression{Lit: t.lit}, nil
	case IDENT:
		p.advance()
		return &variableExpression{Lit: t.lit}, nil
	case STRING_LITERAL:
		p.advance()
		return &stringExpression{Lit: t.lit}, nil
	case '(':
		p.advance()
		sub, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if err := p.expect(')'); err != nil {
			return nil, err
		}
		return &parenExpression{SubExpr: sub}, nil
	default:
		return nil, p.unexpected("number, identifier, string or '('")
	}
}
