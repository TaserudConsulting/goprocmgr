package main // import "github.com/TaserudConsulting/goprocmgr"

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type Runner struct {
	config          *Config
	ActiveProcesses map[string]*ActiveRunner
}

type ActiveRunner struct {
	Cmd    *exec.Cmd
	Stdout []string
	Stderr []string
	Port   uint
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

	if runner.config.Servers[name].UseDirenv {
		if len(splitCmd) > 1 {
			cmd = exec.Command(
				"direnv",
				"exec",
				".",
				splitCmd[0],
				"--",
				splitCmd[1],
			)
		} else {
			cmd = exec.Command(
				"direnv",
				"exec",
				".",
				runner.config.Servers[name].Command,
			)
		}
	} else {
		if len(splitCmd) > 1 {
			cmd = exec.Command(splitCmd[0], splitCmd[1])
		} else {
			cmd = exec.Command(runner.config.Servers[name].Command)
		}
	}

	// Store my active processes
	activeRunner := ActiveRunner{Cmd: cmd}

	// Specify runtime directory.
	cmd.Dir = runner.config.Servers[name].Directory

	// Randomize a port to supply as environment variable.
	port, err := runner.randomizePortNumber()
	if err != nil {
		return err
	}

	// Set environment for running command, first inherit the env from the running command.
	cmd.Env = os.Environ()

	// Then, add the environment variables from the config.
	cmd.Env = append(cmd.Env, fmt.Sprintf("PORT=%d", port))
	cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s", runner.config.Servers[name].Environment["PATH"]))

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
			activeRunner.Stderr = append(activeRunner.Stderr, stderrScanner.Text())
		}
	}()

	// Store the Cmd process as an active process
	runner.ActiveProcesses[name] = &activeRunner

	return nil
}

func (runner *Runner) randomizePortNumber() (uint, error) {
	portRangeSize := int(runner.config.Settings.PortRangeMax - runner.config.Settings.PortRangeMin)

	if len(runner.ActiveProcesses) >= portRangeSize {
		return 0, fmt.Errorf("Out of ports, won't be able to find a port in configured range")
	}

	// Set a changing seed
	rand.Seed(time.Now().UnixNano())

	// Randomize ports within the range
	randomPorts := rand.Perm(portRangeSize)

	for _, randomPort := range randomPorts {
		isPortInUse := false

		randomPort += int(runner.config.Settings.PortRangeMin)

		for _, activeProcess := range runner.ActiveProcesses {
			if int(activeProcess.Port) == randomPort {
				isPortInUse = true
				break
			}
		}

		if !isPortInUse {
			return uint(randomPort), nil
		}
	}

	return 0, fmt.Errorf("Tried to randomize an unused port, failed")
}

func (runner *Runner) Stop(name string) error {
	// Init active processes map
	if runner.ActiveProcesses == nil {
		runner.ActiveProcesses = make(map[string]*ActiveRunner)
	}

	// If server isn't running, just abort.
	if _, ok := runner.ActiveProcesses[name]; !ok {
		return nil
	}

	// Add a go routine to check if the process is killed or not after
	// we've told it to SIGTERM. If it's still running, send a SIGKILL
	// instead to clean up.
	go func() {
		time.Sleep(60 * time.Second)

		if _, ok := runner.ActiveProcesses[name]; ok {
			log.Printf("Force killed the process since it was still alive after 60 seconds %s", name)

			runner.ActiveProcesses[name].Cmd.Process.Kill()

			// Delete old status for process
			delete(runner.ActiveProcesses, name)
		}
	}()

	// Send SIGTERM to the process
	runner.ActiveProcesses[name].Cmd.Process.Signal(syscall.SIGTERM)

	// Wait for process to end
	runner.ActiveProcesses[name].Cmd.Wait()

	// Delete old status for process
	delete(runner.ActiveProcesses, name)

	return nil
}
