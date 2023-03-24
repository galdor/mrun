package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type RunnerCfg struct {
	NbInstances int

	ProgramName      string
	ProgramArguments []string
}

type Runner struct {
	Cfg RunnerCfg

	programs []*Program

	wg sync.WaitGroup
}

func NewRunner(cfg RunnerCfg) *Runner {
	r := Runner{
		Cfg: cfg,
	}

	return &r
}

func (r *Runner) Start() error {
	r.programs = make([]*Program, r.Cfg.NbInstances)

	name := r.Cfg.ProgramName
	args := r.Cfg.ProgramArguments

	for i := 0; i < r.Cfg.NbInstances; i++ {
		program := NewProgram(i+1, name, args)

		if err := program.Start(&r.wg); err != nil {
			for j := 0; j < i; j++ {
				r.programs[j].Kill()
			}

			return fmt.Errorf("cannot start %s: %v", name, err)
		}

		r.programs[i] = program
	}

	return nil
}

func (r *Runner) WaitForTermination() {
	doneChan := make(chan struct{})

	go func() {
		r.wg.Wait()
		close(doneChan)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer close(sigChan)

	select {
	case signo := <-sigChan:
		fmt.Fprintln(os.Stderr)
		r.Info("received signal %d (%v)", signo, signo)
		r.Info("waiting for children")
		<-doneChan

	case <-doneChan:
		r.Info("children terminated")
	}
}

func (r *Runner) Info(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func (r *Runner) Error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
}
