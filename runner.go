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

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	Output    string    `json:"output"`
}

type ActiveRunner struct {
	Cmd  *exec.Cmd
	Port uint
	Logs []LogEntry
}

func (runner *Runner) Start(name string, serve *Serve) error {
	// Init active processes map
	if runner.ActiveProcesses == nil {
		runner.ActiveProcesses = make(map[string]*ActiveRunner)
	}

	if _, ok := runner.config.Servers[name]; !ok {
		return fmt.Errorf("unknown server %s", name)
	}

	if _, ok := runner.ActiveProcesses[name]; ok {
		return fmt.Errorf("server is already running: %s", name)
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

	// Store the port in the runner to expose in the API.
	activeRunner.Port = port

	// Set environment for running command, first inherit the env from the running command.
	cmd.Env = os.Environ()

	// Then, add the environment variables from the config.
	cmd.Env = append(cmd.Env, fmt.Sprintf("PORT=%d", port))
	cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s", runner.config.Servers[name].Environment["PATH"]))

	// Set up pipe to read stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to set up stdout pipe: %s", err)
	}

	// Set up pipe to read stderr
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to set up stderr pipe: %s", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %s", err)
	}

	// Read stdout while the command is executed
	stdoutScanner := bufio.NewScanner(stdout)
	go func() {
		for stdoutScanner.Scan() {
			// Log it to the common log
			activeRunner.Logs = append(activeRunner.Logs, LogEntry{
				Timestamp: time.Now(),
				Message:   stdoutScanner.Text(),
				Output:    "stdout",
			})

			serve.stateChange <- true
		}
	}()

	// Read stderr while the command is executed
	stderrScanner := bufio.NewScanner(stderr)
	go func() {
		for stderrScanner.Scan() {
			// Log it to the common log
			activeRunner.Logs = append(activeRunner.Logs, LogEntry{
				Timestamp: time.Now(),
				Message:   stderrScanner.Text(),
				Output:    "stderr",
			})

			serve.stateChange <- true
		}
	}()

	// Store the Cmd process as an active process
	runner.ActiveProcesses[name] = &activeRunner

	// Notify state change on start
	serve.stateChange <- true

	return nil
}

func (runner *Runner) randomizePortNumber() (uint, error) {
	portRangeSize := int(runner.config.Settings.PortRangeMax - runner.config.Settings.PortRangeMin)

	if len(runner.ActiveProcesses) >= portRangeSize {
		return 0, fmt.Errorf("out of ports, won't be able to find a port in configured range")
	}

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

	return 0, fmt.Errorf("tried to randomize an unused port, failed")
}

func (runner *Runner) Stop(name string, serve *Serve) error {
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

	// Notify state change on stop
	serve.stateChange <- true

	return nil
}
