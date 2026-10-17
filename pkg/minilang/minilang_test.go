package minilang_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"sokoide.com/ll/pkg/minilang"
)

func TestPublicAPI(t *testing.T) {
	program, err := minilang.Parse(`a=a+1;put("a=",a);`)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		t.Run("independent execution", func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			if err := program.Execute(&output); err != nil {
				t.Fatal(err)
			}
			if got := output.String(); got != "a=1\n" {
				t.Fatalf("got %q", got)
			}
		})
	}
}

func TestPublicAPIErrors(t *testing.T) {
	if _, err := minilang.Parse(`put(1);@`); err == nil {
		t.Fatal("expected parse error")
	}
	program, err := minilang.Parse(`put(1/0);`)
	if err != nil {
		t.Fatal(err)
	}
	if err := program.Execute(io.Discard); err == nil {
		t.Fatal("expected evaluation error")
	}
	if err := program.Execute(nil); err == nil {
		t.Fatal("expected nil writer error")
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestOutputError(t *testing.T) {
	for _, source := range []string{`put("text");`, `put(001);`, `put(a);`, `put(1+1);`, `put("");`} {
		program, err := minilang.Parse(source)
		if err != nil {
			t.Fatal(err)
		}
		if err := program.Execute(failingWriter{}); !errors.Is(err, io.ErrClosedPipe) {
			t.Fatalf("%s: got %v", source, err)
		}
	}
}
