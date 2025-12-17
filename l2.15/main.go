package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

var builtins = map[string]func([]string) error{
	"cd":  builtinCd,
	"pwd": builtinPwd,
	"echo": builtinEcho,
	"kill": builtinKill,
	"ps":  builtinPs,
}

func builtinCd(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("cd: expected argument")
	}
	return os.Chdir(args[1])
}

func builtinPwd(args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Println(dir)
	return nil
}

func builtinEcho(args []string) error {
	if len(args) > 1 {
		fmt.Println(strings.Join(args[1:], " "))
	} else {
		fmt.Println()
	}
	return nil
}

func builtinKill(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("kill: expected PID")
	}
	pid, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("kill: invalid PID")
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Signal(syscall.SIGTERM)
}

func builtinPs(args []string) error {
	cmd := exec.Command("ps", "aux")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		for range signalChan {
		}
	}()

	fmt.Println("Simple Shell (type 'exit' or Ctrl+D to quit)")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		if isStdinTerminal() {
    		fmt.Print("$ ")
		}

		if !scanner.Scan() {
			fmt.Println()
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		pipeStages := splitPipeline(line)
		if len(pipeStages) == 0 {
			continue
		}

		if err := runPipeline(pipeStages); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "read error: %v\n", err)
	}
}

func splitPipeline(line string) []string {
	var stages []string
	var current strings.Builder
	inQuotes := false
	escaped := false

	for _, ch := range line {
		if escaped {
			current.WriteRune(ch)
			escaped = false
			continue
		}

		if ch == '\\' {
			escaped = true
			continue
		}

		if ch == '"' || ch == '\'' {
			inQuotes = !inQuotes
			current.WriteRune(ch)
			continue
		}

		if ch == '|' && !inQuotes {
			stages = append(stages, strings.TrimSpace(current.String()))
			current.Reset()
		} else {
			current.WriteRune(ch)
		}
	}

	stages = append(stages, strings.TrimSpace(current.String()))
	return stages
}

func runPipeline(stages []string) error {
	if len(stages) == 0 {
		return nil
	}

	var err error
	var prevStdout io.Reader = nil

	for i, stage := range stages {
		if stage == "" {
			continue
		}

		args := parseArgs(stage)
		if len(args) == 0 {
			continue
		}

		cmdName := args[0]

		if i == 0 && builtins[cmdName] != nil {
			if len(stages) > 1 {
				return fmt.Errorf("builtin command '%s' cannot be used in a pipeline", cmdName)
			}
			return builtins[cmdName](args)
		}

		cmd := exec.Command(cmdName, args[1:]...)

		if i == 0 {
			cmd.Stdin = os.Stdin
		} else {
			cmd.Stdin = prevStdout
		}

		var pipeOut io.ReadCloser
		if i == len(stages)-1 {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		} else {
			pipeOut, err = cmd.StdoutPipe()
			if err != nil {
				return err
			}
		}

		if err = cmd.Start(); err != nil {
			if pipeOut != nil {
				pipeOut.Close()
			}
			return err
		}

		if i < len(stages)-1 {
			prevStdout = pipeOut
		}

		if err = cmd.Wait(); err != nil {
			return err
		}

		if pipeOut != nil {
			pipeOut.Close()
		}
	}

	return nil
}

func parseArgs(line string) []string {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	return strings.Fields(line)
}

func isStdinTerminal() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}