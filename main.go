package main

import (
	"flag"
	"io"
	"os"

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
	s := new(scanner)
	s.Init(string(source))
	program, err := parse(s)
	if err != nil {
		log.Errorf("Parse error: %v", err)
		os.Exit(1)
	}
	if _, err := evaluateStmts(program); err != nil {
		log.Errorf("Evaluation error: %v", err)
		os.Exit(1)
	}
}
