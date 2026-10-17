package minilang

import "strconv"

// ========================================
const (
	// EOF end of file
	EOF = -1
)

// Named tokens do not overlap with punctuation rune values.
const (
	ILLEGAL = 256 + iota
	PUT
	WHILE
	EQOP
	NEOP
	GEOP
	GTOP
	LEOP
	LTOP
	NUMBER_LITERAL
	IDENT
	STRING_LITERAL
)

var keywords = map[string]int{
	"put":   PUT,
	"while": WHILE,
	"==":    EQOP,
	"!=":    NEOP,
	">=":    GEOP,
	">":     GTOP,
	"<=":    LEOP,
	"<":     LTOP,
}

// tokenType enum
type tokenType int

const (
	tokenTypeNone tokenType = iota
	tokenTypeNumberLiteral
	tokenTypeStringLiteral
	tokenTypeIdent
	tokenTypeKeyword
	tokenTypeOp
)

type token struct {
	tok int
	lit string
	pos position
	tt  tokenType
}

type position struct {
	Line   int
	Column int
}

type scanner struct {
	src      []rune
	offset   int
	lineHead int
	line     int
}

func (s *scanner) Init(src string) {
	*s = scanner{src: []rune(src)}
}

func (s *scanner) Scan() (tok int, lit string, pos position, tt tokenType) {
	s.skipWhiteSpace()
	pos = s.position()
	switch ch := s.peek(); {
	case isDigit(ch):
		tt = tokenTypeNumberLiteral
		tok, lit = NUMBER_LITERAL, s.scanNumber()
	case isLetter(ch):
		tt = tokenTypeIdent
		tok, lit = IDENT, s.scanIdentifier()
		if t, ok := keywords[lit]; ok {
			tok = t
			tt = tokenTypeKeyword
		}
	case ch == '"':
		s.next() // Skip the opening quote.
		lit = s.scanString('"')
		tok = STRING_LITERAL
		tt = tokenTypeStringLiteral
		if s.peek() != -1 {
			s.next()
		} else {
			tok = ILLEGAL
		}
	case ch == '\'':
		s.next() // Skip the opening quote.
		lit = s.scanString('\'')
		tok = STRING_LITERAL
		tt = tokenTypeStringLiteral
		if s.peek() != -1 {
			s.next()
		} else {
			tok = ILLEGAL
		}
	case ch == '<':
		s.next()
		if s.peek() == '=' {
			s.next()
			lit = "<="
		} else {
			lit = "<"
		}
		tok = keywords[lit]
		tt = tokenTypeOp
	case ch == '>':
		s.next()
		if s.peek() == '=' {
			s.next()
			lit = ">="
		} else {
			lit = ">"
		}
		tok = keywords[lit]
		tt = tokenTypeOp
	case ch == '=':
		s.next()
		if s.peek() == '=' {
			s.next()
			lit = "=="
			tok = keywords[lit]
			tt = tokenTypeOp
		} else {
			tok = int(ch)
			lit = "="
			tt = tokenTypeOp
		}
	case ch == '!':
		s.next()
		if s.peek() == '=' {
			s.next()
			lit = "!="
			tok = keywords[lit]
			tt = tokenTypeOp
		} else {
			// `!` alone is not part of the language
			tok = ILLEGAL
			lit = "!"
		}
	default:
		switch ch {
		case -1:
			tok = EOF
		case ',', '.', '[', ']', '(', ')', '{', '}', ';', '+', '-', '*', '/', '%':
			tok = int(ch)
			lit = string(ch)
			tt = tokenTypeOp
		default:
			// Invalid input must remain visible to the parser.
			tok = ILLEGAL
			lit = string(ch)
		}
		s.next()
	}
	return
}

// ========================================

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func isWhiteSpace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func (s *scanner) peek() rune {
	if !s.reachEOF() {
		return s.src[s.offset]
	}
	return -1
}

func (s *scanner) next() {
	if !s.reachEOF() {
		if s.peek() == '\n' {
			s.lineHead = s.offset + 1
			s.line++
		}
		s.offset++
	}
}

func (s *scanner) reachEOF() bool {
	return len(s.src) <= s.offset
}

func (s *scanner) position() position {
	return position{Line: s.line + 1, Column: s.offset - s.lineHead + 1}
}

func (s *scanner) skipWhiteSpace() {
	for isWhiteSpace(s.peek()) {
		s.next()
	}
}

func (s *scanner) scanIdentifier() string {
	var ret []rune
	for isLetter(s.peek()) || isDigit(s.peek()) {
		ret = append(ret, s.peek())
		s.next()
	}
	return string(ret)
}

func (s *scanner) scanNumber() string {
	var ret []rune
	for isDigit(s.peek()) {
		ret = append(ret, s.peek())
		s.next()
	}
	return string(ret)
}

func (s *scanner) scanString(endingChar rune) string {
	var ret []rune
	for s.peek() != endingChar && s.peek() != -1 {
		ret = append(ret, s.peek())
		s.next()
	}
	return string(ret)
}

// tokenName maps internal token numbers to grammar terminals for diagnostics.
func tokenName(tok int) string {
	switch tok {
	case EOF:
		return "EOF"
	case ILLEGAL:
		return "ILLEGAL"
	case IDENT:
		return "IDENT"
	case NUMBER_LITERAL:
		return "NUMBER"
	case STRING_LITERAL:
		return "STRING"
	case PUT:
		return "put"
	case WHILE:
		return "while"
	case EQOP:
		return "=="
	case NEOP:
		return "!="
	case GEOP:
		return ">="
	case GTOP:
		return ">"
	case LEOP:
		return "<="
	case LTOP:
		return "<"
	default:
		return strconv.QuoteRune(rune(tok))
	}
}
