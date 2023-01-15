package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

type Runner struct {
	config          *Config
	ActiveProcesses map[string]*ActiveRunner
}

type ActiveRunner struct {
	Cmd    *exec.Cmd
	Stdout []string
	Stderr []string
}

func (runner *Runner) Start(name string) error {
	// Init active processes map
	if runner.ActiveProcesses == nil {
		runner.ActiveProcesses = make(map[string]*ActiveRunner)
	}

	if _, ok := runner.config.Servers[name]; !ok {
		return fmt.Errorf("Unknown server %s", name)
	}

	if _, ok := runner.ActiveProcesses[name]; ok {
		return fmt.Errorf("Server is already running: %s", name)
	}

	// Split the command on the first space since exec.Command will
	// look for the first argument only in the path as a binary name.
	splitCmd := strings.SplitN(runner.config.Servers[name].Command, " ", 2)

	// Set up a command.
	var cmd *exec.Cmd
	if len(splitCmd) > 1 {
		cmd = exec.Command(splitCmd[0], splitCmd[1])
	} else {
		cmd = exec.Command(runner.config.Servers[name].Command)
	}

	// Store my active processes
	activeRunner := ActiveRunner{Cmd: cmd}

	// Specify runtime directory.
	cmd.Dir = runner.config.Servers[name].Directory

	// Set up pipe to read stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Failed to set up stdout pipe: %s", err)
	}

	// Set up pipe to read stderr
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("Failed to set up stderr pipe: %s", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("Failed to start process: %s", err)
	}

	// Read stdout while the command is executed
	stdoutScanner := bufio.NewScanner(stdout)
	go func() {
		for stdoutScanner.Scan() {
			activeRunner.Stdout = append(activeRunner.Stdout, stdoutScanner.Text())
		}
	}()

	// Read stderr while the command is executed
	stderrScanner := bufio.NewScanner(stderr)
	go func() {
		for stderrScanner.Scan() {
			activeRunner.Stderr = append(activeRunner.Stderr, stdoutScanner.Text())
		}
	}()

	// Store the Cmd process as an active process
	runner.ActiveProcesses[name] = &activeRunner

	return nil
}
