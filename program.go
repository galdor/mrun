package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
)

type Program struct {
	Id        int
	Name      string
	Arguments []string

	cmd *exec.Cmd

	wg       *sync.WaitGroup
	outputWg sync.WaitGroup
}

func NewProgram(id int, name string, args []string) *Program {
	p := &Program{
		Id:        id,
		Name:      name,
		Arguments: args,
	}

	return p
}

func (p *Program) Start(wg *sync.WaitGroup) error {
	p.wg = wg

	p.cmd = exec.Command(p.Name, p.Arguments...)

	stdoutPipe, err := p.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("cannot create stdout pipe: %w", err)
	}

	stderrPipe, err := p.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("cannot create stderr pipe: %w", err)
	}

	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("cannot start program: %w", err)
	}

	p.wg.Add(1)
	go p.main()

	p.outputWg.Add(2)
	go p.readOutput("stdout", stdoutPipe)
	go p.readOutput("stderr", stderrPipe)

	p.Info("program started (pid: %d)", p.cmd.Process.Pid)

	return nil
}

func (p *Program) Kill() error {
	if err := p.cmd.Process.Signal(syscall.SIGKILL); err != nil {
		return fmt.Errorf("cannot send sigkill to instance %d: %w", p.Id, err)
	}

	return nil
}

func (p *Program) Info(format string, args ...interface{}) {
	args2 := append([]interface{}{p.Id}, args...)
	fmt.Printf("[%3d] "+format+"\n", args2...)
}

func (p *Program) Error(format string, args ...interface{}) {
	args2 := append([]interface{}{p.Id}, args...)
	fmt.Fprintf(os.Stderr, "[%3d] error: "+format+"\n", args2...)
}

func (p *Program) main() {
	defer p.wg.Done()

	p.outputWg.Wait()

	err := p.cmd.Wait()
	p.reportResult(err)
}

func (p *Program) readOutput(name string, output io.Reader) {
	defer p.outputWg.Done()

	r := bufio.NewReader(output)

	var buf []byte

	for {
		line, isPrefix, err := r.ReadLine()
		if err == io.EOF {
			if len(buf) > 0 {
				p.Info("%s: %s", name, buf)
			}

			break
		} else if err != nil {
			p.Error("cannot read %s: %v", name, err)
			return
		}

		buf = append(buf, line...)

		if isPrefix {
			continue
		}

		p.Info("%s: %s", name, buf)

		buf = []byte{}
	}
}

func (p *Program) reportResult(result error) {
	var exitErr *exec.ExitError

	if result == nil {
		p.Info("program exited successfully")
	} else if errors.As(result, &exitErr) {
		if s, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			if s.Exited() {
				p.Error("program exited with status %d", s.ExitStatus())
			} else if s.Signaled() {
				signo := s.Signal()
				p.Error("program killed by signal %d (%v)", signo, signo)
			} else {
				p.Error("program failed: %v", exitErr)
			}
		} else {
			s := exitErr.ProcessState

			if s.Exited() {
				p.Error("program exited with status %d", s.ExitCode())
			} else {
				p.Error("program failed: %v", exitErr)
			}
		}
	}
}
