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
	flag.Parse()
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
	program, err := minilang.Parse(string(source))
	if err != nil {
		log.Errorf("Parse error: %v", err)
		os.Exit(1)
	}
	if err := program.Execute(os.Stdout); err != nil {
		log.Errorf("Evaluation error: %v", err)
		os.Exit(1)
	}
}
