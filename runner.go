package main

import (
	"bytes"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"text/template"
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

	for i := 0; i < r.Cfg.NbInstances; i++ {
		instanceId := i + 1

		args := make([]string, len(r.Cfg.ProgramArguments))
		for idx, arg := range r.Cfg.ProgramArguments {
			arg2, err := r.renderArgument(arg, instanceId)
			if err != nil {
				return fmt.Errorf("cannot render argument %q: %w", arg, err)
			}

			args[idx] = arg2
		}

		program := NewProgram(instanceId, name, args)

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

func (r *Runner) renderArgument(arg string, instanceId int) (string, error) {
	ctx := struct {
		InstanceId int
	}{
		InstanceId: instanceId,
	}

	tpl, err := template.New("string").Parse(arg)
	if err != nil {
		return "", fmt.Errorf("cannot parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, ctx); err != nil {
		return "", fmt.Errorf("cannot render template: %w", err)
	}

	return buf.String(), nil
}
