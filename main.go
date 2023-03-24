package main

import (
	"strconv"

	"github.com/exograd/go-program"
)

func main() {
	p := program.NewProgram("mrun",
		"utility to execute multiple instances of a program in parallel")

	p.AddArgument("n", "the number of instances to run")
	p.AddArgument("program", "the name of the program to run")
	p.AddTrailingArgument("argument",
		"a list of arguments to pass to the command")

	p.SetMain(cmdMain)
	p.ParseCommandLine()
	p.Run()
}

func cmdMain(p *program.Program) {
	nString := p.ArgumentValue("n")
	n, err := strconv.ParseInt(nString, 10, 64)
	if err != nil || n < 1 {
		p.Fatal("invalid number of instances %q", nString)
	}

	cfg := RunnerCfg{
		NbInstances: int(n),

		ProgramName:      p.ArgumentValue("program"),
		ProgramArguments: p.TrailingArgumentValues("argument"),
	}

	runner := NewRunner(cfg)

	if err := runner.Start(); err != nil {
		p.Fatal("cannot start runner: %w", err)
	}

	runner.WaitForTermination()
}
