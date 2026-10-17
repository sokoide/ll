package main

import (
	"strings"
	"testing"
)

func parseSource(source string) ([]expression, error) {
	s := new(scanner)
	s.Init(source)
	return parse(s)
}

func TestEvaluation(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		want         int
	}{
		{"precedence", `a=1+5*2+9/3;`, 14},
		{"parentheses", `a=(40+1)+2/2;`, 42},
		{"left subtraction", `a=20-5-3;`, 12},
		{"left division", `a=24/4/2;`, 3},
		{"modulo", `a=20%6*2;`, 4},
		{"while", `a=0;while(a<5){a=a+1;b=b+1;};`, 5},
		{"truth condition", `a=3;while(a){a=a-1;}`, 0},
		{"negative condition", `a=0;while(0-1){a=9;}`, 0},
		{"undefined", `a=x+1;`, 1},
		{"nested scopes", `a=0;while(a<2){b=0;while(b<3){a=a+1;b=b+1;}}`, 3},
		{"CRLF", "a=1;\r\na=a+1;", 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			scopeStack = NewScopeStack()
			program, err := parseSource(tt.source)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = evaluateStmts(program); err != nil {
				t.Fatal(err)
			}
			if got, _ := scopeStack.Get("a"); got != tt.want {
				t.Fatalf("a=%d, want %d", got, tt.want)
			}
			if _, scope := scopeStack.Get("b"); scope != nil {
				t.Fatal("loop-local b leaked")
			}
		})
	}
}

func TestComparisons(t *testing.T) {
	for _, cond := range []string{"1==1", "1!=2", "2>=2", "2>1", "1<=1", "1<2", "2==1", "1!=1", "1>=2", "1>2", "2<=1", "2<1"} {
		t.Run(cond, func(t *testing.T) {
			scopeStack = NewScopeStack()
			program, err := parseSource("while(" + cond + "){;}")
			if err != nil {
				t.Fatal(err)
			}
			got, err := evaluateExpr(program[0].(*whileStatement).Cond)
			if err != nil {
				t.Fatal(err)
			}
			want := 0
			for _, truth := range []string{"1==1", "1!=2", "2>=2", "2>1", "1<=1", "1<2"} {
				if cond == truth {
					want = 1
				}
			}
			if got != want {
				t.Fatalf("got %d, want %d", got, want)
			}
		})
	}
}

func TestSyntaxErrors(t *testing.T) {
	for _, src := range []string{`put(-1);`, `put();`, `put(1,);`, `a=1`, `while(1){`, `}`, `{;}`, `1+2;`, `a=1<2;`, `put(1<2);`, `while(1<2<3){}`, `while((1<2)){}`, `put(1);@put(2);`, `put(1);!`, `put(1);` + "\x00", `put("abc);`, `put('abc);`, `put((1);`, `while(1)put(1);`, `put(1);extra`} {
		t.Run(src, func(t *testing.T) {
			if _, err := parseSource(src); err == nil {
				t.Fatal("expected syntax error")
			}
		})
	}
}

func TestEvaluationErrors(t *testing.T) {
	for _, tt := range []struct{ src, want string }{
		{`a="x"+"y";`, "not supported in arithmetic"},
		{`a="x";`, "not supported in arithmetic"},
		{`put(("x"));`, "not supported in arithmetic"},
		{`a=5/0;`, "division by zero"},
		{`a=5%0;`, "modulo by zero"},
		{`while("x"){}`, "not supported in arithmetic"},
		{`while(1){a=1/0;}`, "division by zero"},
	} {
		t.Run(tt.src, func(t *testing.T) {
			scopeStack = NewScopeStack()
			outer := scopeStack.current
			program, err := parseSource(tt.src)
			if err != nil {
				t.Fatal(err)
			}
			_, err = evaluateStmts(program)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("got %v, want %s", err, tt.want)
			}
			if scopeStack.current != outer {
				t.Fatal("scope not restored after error")
			}
		})
	}
}

func TestEmptyAndStrings(t *testing.T) {
	for _, src := range []string{"", ";;;", `while(0){}`, `put("",'日本語',001);`} {
		if _, err := parseSource(src); err != nil {
			t.Fatal(err)
		}
	}
}

func TestErrorPosition(t *testing.T) {
	_, err := parseSource("put(1);\n  @")
	if err == nil || !strings.Contains(err.Error(), "Line:2 Column:3") {
		t.Fatalf("got %v", err)
	}
}

func TestScannerReinitialization(t *testing.T) {
	s := new(scanner)
	s.Init("a=1;\n")
	if _, err := parse(s); err != nil {
		t.Fatal(err)
	}
	s.Init("b=2;")
	if _, err := parse(s); err != nil {
		t.Fatal(err)
	}
}
