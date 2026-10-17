package minilang

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestASTShape(t *testing.T) {
	for _, tt := range []struct{ source, want string }{
		{`put(20-5-3);`, "Put\n  Binary '-'\n    Binary '-'\n      Number 20\n      Number 5\n    Number 3\n"},
		{`put(1+2*3);`, "Put\n  Binary '+'\n    Number 1\n    Binary '*'\n      Number 2\n      Number 3\n"},
		{`put((1+2)*3);`, "Put\n  Binary '*'\n    Paren\n      Binary '+'\n        Number 1\n        Number 2\n    Number 3\n"},
	} {
		t.Run(tt.source, func(t *testing.T) {
			p, err := Parse(tt.source)
			if err != nil {
				t.Fatal(err)
			}
			var got bytes.Buffer
			if err := p.WriteAST(&got); err != nil {
				t.Fatal(err)
			}
			if got.String() != tt.want {
				t.Fatalf("AST:\n%s\nwant:\n%s", got.String(), tt.want)
			}
		})
	}
}

func TestWhileAST(t *testing.T) {
	p, err := Parse(`while(a<3){a=a+1;}`)
	if err != nil {
		t.Fatal(err)
	}
	var got bytes.Buffer
	if err := p.WriteAST(&got); err != nil {
		t.Fatal(err)
	}
	want := "While\n  Condition\n    Compare <\n      Variable a\n      Number 3\n  Body\n    Assign a\n      Binary '+'\n        Variable a\n        Number 1\n"
	if got.String() != want {
		t.Fatalf("got:\n%s", got.String())
	}
}

func TestTokens(t *testing.T) {
	var got bytes.Buffer
	if err := WriteTokens(`put(42);`, &got); err != nil {
		t.Fatal(err)
	}
	want := "1:1 put      \"put\"\n1:4 '('      \"(\"\n1:5 NUMBER   \"42\"\n1:7 ')'      \")\"\n1:8 ';'      \";\"\n1:9 EOF      \"\"\n"
	if got.String() != want {
		t.Fatalf("got:\n%s", got.String())
	}
	// Maximal munch must distinguish reserved words from longer identifiers.
	s := new(scanner)
	s.Init("while while1 <= < != == =\r\n'日本語'")
	for _, want := range []int{WHILE, IDENT, LEOP, LTOP, NEOP, EQOP, '=', STRING_LITERAL, EOF} {
		got, _, _, _ := s.Scan()
		if got != want {
			t.Fatalf("got %s want %s", tokenName(got), tokenName(want))
		}
	}
	for _, src := range []string{"@", "!", `"unclosed`} {
		if err := WriteTokens(src, io.Discard); err == nil {
			t.Fatalf("accepted %q", src)
		}
	}
}

func TestReadableErrors(t *testing.T) {
	for _, tt := range []struct{ source, want string }{
		{`a=1`, "expected ';', got EOF"},
		{`a=1 2;`, "expected ';', got NUMBER"},
		{`while(1){`, "expected '}', got EOF"},
		{`put(1+);`, "expected number, identifier or '('"},
	} {
		_, err := Parse(tt.source)
		if err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Fatalf("%s: got %v want %q", tt.source, err, tt.want)
		}
	}
}

func TestBytecode(t *testing.T) {
	p, err := Parse(`put(1+2*3);`)
	if err != nil {
		t.Fatal(err)
	}
	code, err := p.Compile()
	if err != nil {
		t.Fatal(err)
	}
	var got bytes.Buffer
	if err := code.Disassemble(&got); err != nil {
		t.Fatal(err)
	}
	want := "0000 PUSH 1\n0001 PUSH 2\n0002 PUSH 3\n0003 BINARY '*'\n0004 BINARY '+'\n0005 PRINT_INT\n0006 NEWLINE\n"
	if got.String() != want {
		t.Fatalf("got:\n%s", got.String())
	}
}

func TestBytecodeWhile(t *testing.T) {
	p, err := Parse(`a=0;while(a<2){a=a+1;}put(a);`)
	if err != nil {
		t.Fatal(err)
	}
	b, err := p.Compile()
	if err != nil {
		t.Fatal(err)
	}
	var got bytes.Buffer
	if err := b.Disassemble(&got); err != nil {
		t.Fatal(err)
	}
	want := "0000 PUSH 0\n0001 STORE a\n0002 ENTER_SCOPE\n0003 LOAD a\n0004 PUSH 2\n0005 BINARY <\n0006 JUMP_IF_FALSE 12\n0007 LOAD a\n0008 PUSH 1\n0009 BINARY '+'\n0010 STORE a\n0011 JUMP 3\n0012 LEAVE_SCOPE\n0013 LOAD a\n0014 PRINT_INT\n0015 NEWLINE\n"
	if got.String() != want {
		t.Fatalf("got:\n%s", got.String())
	}
}

func TestExamples(t *testing.T) {
	files, err := filepath.Glob("../../examples/*.mini")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no examples")
	}
	for _, path := range files {
		t.Run(filepath.Base(path), func(t *testing.T) {
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(strings.TrimSuffix(path, ".mini") + ".out")
			if err != nil {
				t.Fatal(err)
			}
			p, err := Parse(string(source))
			if err != nil {
				t.Fatal(err)
			}
			b, err := p.Compile()
			if err != nil {
				t.Fatal(err)
			}
			for _, run := range []func(io.Writer) error{p.Execute, b.Execute} {
				var got bytes.Buffer
				if err := run(&got); err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(got.Bytes(), want) {
					t.Fatalf("got %q want %q", got.String(), want)
				}
			}
		})
	}
}

func TestVMHandwrittenInstructions(t *testing.T) {
	code := &Bytecode{code: []instruction{
		{op: opPush, text: "20"}, {op: opPush, text: "5"},
		{op: opBinary, arg: '-'}, {op: opPrintInt},
	}}
	var got bytes.Buffer
	if err := code.Execute(&got); err != nil {
		t.Fatal(err)
	}
	if got.String() != "15" {
		t.Fatal(got.String())
	}
}

func TestREADMEGrammar(t *testing.T) {
	readme, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatal(err)
	}
	grammar, err := os.ReadFile("../../grammar.ebnf")
	if err != nil {
		t.Fatal(err)
	}
	text := string(grammar)
	syntax := strings.TrimSpace(text[strings.Index(text, "program "):strings.Index(text, "(* Lexical rules:")])
	if !strings.Contains(string(readme), "```ebnf\n"+syntax+"\n```") {
		t.Fatal("README EBNF differs from grammar.ebnf")
	}
	for _, ext := range []string{"mini", "out"} {
		fixture, err := os.ReadFile("../../examples/readme." + ext)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(readme), strings.TrimSpace(string(fixture))) {
			t.Fatalf("README is missing readme.%s", ext)
		}
	}
}

func TestVMStackAndJumpErrors(t *testing.T) {
	for _, tt := range []struct {
		code []instruction
		want string
	}{
		{[]instruction{{op: opBinary, arg: '-'}}, "operand stack underflow"},
		{[]instruction{{op: opJump, arg: -1}}, "invalid jump target"},
		{[]instruction{{op: opPush, text: "1"}}, "unbalanced operand stack"},
	} {
		err := (&Bytecode{code: tt.code}).Execute(io.Discard)
		if err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Fatalf("got %v want %s", err, tt.want)
		}
	}
}
