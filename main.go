package main

import (
	"github.com/exograd/go-program"
)

func main() {
	p := program.NewProgram("mrun",
		"utility to execute multiple instances of a program in parallel")

	p.SetMain(cmdMain)
	p.ParseCommandLine()
	p.Run()
}

func cmdMain(p *program.Program) {
}
