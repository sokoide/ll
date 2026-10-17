package minilang_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"sokoide.com/ll/pkg/minilang"
)

func TestVMMatchesTree(t *testing.T) {
	cases := []struct{ source, want string }{
		{"", ""},
		{`;;;put("ok");`, "ok\n"},
		{`put(1+2*3,(1+2)*3,20-5-3,24/4/2,20%6*2);`, "791234\n"},
		{`put("",'日本語',001,(001));`, "日本語0011\n"},
		{`put(x);a=x+1;put(a);`, "0\n1\n"},
		{`a=0;while(a<3){a=a+1;b=b+1;put(b);}put(a,b);`, "1\n2\n3\n30\n"},
		{`a=0;while(a<2){b=0;while(b<2){put(a,b);b=b+1;}a=a+1;}put(b);`, "00\n01\n10\n11\n0\n"},
		{`a=3;while(a){a=a-1;}while(0-1){put("bad");}put(a);`, "0\n"},
	}
	for _, op := range []string{"==", "!=", ">=", ">", "<=", "<"} {
		rhs := "2"
		start := "1"
		step := "a=a+1;"
		want := "1\n2\n"
		switch op {
		case "==":
			rhs = "1"
			want = "1\n2\n"
		case "!=":
			want = "1\n2\n"
		case ">=":
			start = "2"
			rhs = "2"
			step = "a=a-1;"
			want = "2\n1\n"
		case ">":
			start = "2"
			rhs = "1"
			step = "a=a-1;"
			want = "2\n1\n"
		case "<=":
			rhs = "1"
			want = "1\n2\n"
		case "<":
			want = "1\n2\n"
		}
		cases = append(cases, struct{ source, want string }{"a=" + start + ";while(a" + op + rhs + "){put(a);" + step + "}put(a);", want})
	}
	for _, tt := range cases {
		t.Run(tt.source, func(t *testing.T) {
			p, err := minilang.Parse(tt.source)
			if err != nil {
				t.Fatal(err)
			}
			code, err := p.Compile()
			if err != nil {
				t.Fatal(err)
			}
			for _, run := range []func(io.Writer) error{p.Execute, code.Execute, code.Execute} {
				var output bytes.Buffer
				if err := run(&output); err != nil {
					t.Fatal(err)
				}
				if got := output.String(); got != tt.want {
					t.Fatalf("got %q, want %q", got, tt.want)
				}
			}
		})
	}
}

func TestVMErrors(t *testing.T) {
	for _, source := range []string{`put(1/0);`, `a=1%0;`, `while(1){put(1/0);}`, `a=999999999999999999999999999999999;`} {
		p, err := minilang.Parse(source)
		if err != nil {
			t.Fatal(err)
		}
		code, err := p.Compile()
		if err != nil {
			t.Fatal(err)
		}
		treeErr := p.Execute(io.Discard)
		vmErr := code.Execute(io.Discard)
		if treeErr == nil || vmErr == nil || !strings.Contains(vmErr.Error(), treeErr.Error()) {
			t.Fatalf("tree=%v vm=%v", treeErr, vmErr)
		}
	}
	p, _ := minilang.Parse(`put(1);`)
	code, _ := p.Compile()
	if err := code.Execute(failingWriter{}); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal(err)
	}
	if err := code.Execute(nil); err == nil {
		t.Fatal("expected nil writer error")
	}
}

func TestVMConcurrentRuns(t *testing.T) {
	p, _ := minilang.Parse(`a=a+1;put(a);`)
	code, _ := p.Compile()
	for i := 0; i < 4; i++ {
		t.Run("fresh state", func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			if err := code.Execute(&output); err != nil {
				t.Fatal(err)
			}
			if output.String() != "1\n" {
				t.Fatal(output.String())
			}
		})
	}
}
