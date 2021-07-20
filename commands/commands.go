package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/preludeorg/pneuma/commands/pty"
	"github.com/preludeorg/pneuma/util"
	"os"
	"os/exec"
	"strings"
	"time"
)

//RunCommand executes a given command
func RunCommand(message string, executor string, payloadPath string, agent *util.AgentConfig) (string, int, int) {
	switch executor {
	case "keyword":
		task := splitMessage(message, '.')
		switch task[0] {
		case "config":
			return updateConfiguration(task[1], agent)
		case "shell":
			return pty.SpawnShell(task[1], agent)
		case "exit":
			return shutdown(agent)
		default:
			return "Keyword selected not available for agent", 0, 0
		}
	default:
		util.DebugLogf("Running instruction")
		bites, status, pid := execute(message, executor, agent)
		return string(bites), status, pid
	}
}

func execute(command string, executor string, agent *util.AgentConfig) ([]byte, int, int) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(agent.CommandTimeout) * time.Second)
	defer cancel()
	bites, pid, status := execution(getShellCommand(ctx, executor, command))
	if ctx.Err() == context.DeadlineExceeded {
		bites = []byte("Command timed out.")
	}
	return []byte(fmt.Sprintf("%s%s", bites, "\n")), status, pid
}

func execution(command *exec.Cmd) ([]byte, int, int){
	var bites []byte
	var status int
	var pid int
	if out, err := command.Output(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			bites = append(out, exitError.Stderr...)
			pid = exitError.Pid()
			status = exitError.ProcessState.ExitCode()
		} else {
			bites = append(out, []byte(err.Error())...)
			pid = -1
			status = command.ProcessState.ExitCode()
		}
	} else {
		bites = out
		pid = command.ProcessState.Pid()
		status = command.ProcessState.ExitCode()
	}
	return bites, pid, status
}

func updateConfiguration(config string, agent *util.AgentConfig) (string, int, int) {
	var newConfig map[string]interface{}
	err := json.Unmarshal([]byte(config), &newConfig)
	if err == nil {
		agent.SetAgentConfig(newConfig)
		return "Successfully updated agent configuration.", 0, os.Getpid()
	}
	return err.Error(), 1, os.Getpid()
}

func shutdown(agent *util.AgentConfig) (string, int, int) {
	go func(a *util.AgentConfig) {
		time.Sleep(time.Duration(a.KillSleep) * time.Second)
		os.Exit(0)
	}(agent)
	return fmt.Sprintf("Exiting agent in %d seconds", agent.KillSleep), 0, os.Getpid()
}

func splitMessage(message string, splitRune rune) []string {
	quoted := false
	values := strings.FieldsFunc(message, func(r rune) bool {
		if r == '"' {
			quoted = !quoted
		}
		return !quoted && r == splitRune
	})
	return values
}