package main

import (
	"flag"
	"io"
	"os"

	"sokoide.com/ll/pkg/minilang"

	log "github.com/sirupsen/logrus"
)

func main() {
	level := flag.String("logLevel", "debug", "TRACE | DEBUG | INFO | WARN | ERROR")
	engine := flag.String("engine", "tree", "tree | vm")
	dump := flag.String("dump", "", "tokens | ast | bytecode (inspect without executing)")
	flag.Parse()
	if *engine != "tree" && *engine != "vm" {
		log.Errorf("unknown engine %q", *engine)
		os.Exit(1)
	}
	if *dump != "" && *dump != "tokens" && *dump != "ast" && *dump != "bytecode" {
		log.Errorf("unknown dump mode %q", *dump)
		os.Exit(1)
	}
	ll, err := log.ParseLevel(*level)
	if err != nil {
		log.Warnf("logLevel: %s is not supported. falling back to INFO", *level)
		ll = log.InfoLevel
	}
	log.SetLevel(ll)
	source, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	if *dump == "tokens" {
		if err := minilang.WriteTokens(string(source), os.Stdout); err != nil {
			log.Error(err)
			os.Exit(1)
		}
		return
	}
	program, err := minilang.Parse(string(source))
	if err != nil {
		log.Errorf("Parse error: %v", err)
		os.Exit(1)
	}
	if *dump == "ast" {
		err = program.WriteAST(os.Stdout)
	} else if *dump == "bytecode" || *engine == "vm" {
		var code *minilang.Bytecode
		code, err = program.Compile()
		if err == nil {
			if *dump == "bytecode" {
				err = code.Disassemble(os.Stdout)
			} else {
				err = code.Execute(os.Stdout)
			}
		}
	} else {
		err = program.Execute(os.Stdout)
	}
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
