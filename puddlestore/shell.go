package main

import (
	"./puddlestore"
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type Shell struct {
	p    *puddlestore.PuddleNode
	c    *puddlestore.Client
	done chan bool
}

type command struct {
	f     func(shell *Shell, args []string) error
	usage string
	help  string
	args  int
}

func (shell *Shell) interact(commands map[string]command) {
	usage := func() {
		fmt.Println("Commands:")
		for _, cmd := range commands {
			fmt.Printf(" - %-25s %s\n", cmd.usage, cmd.help)
		}
		fmt.Println(" - exit")
	}

	usage()

	in := bufio.NewReader(os.Stdin)
LOOP:
	for {
		fmt.Print("> ")
		input, err := in.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			continue
		}
		// Trim trailing newline
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		args := strings.Split(input, " ")
		switch args[0] {
		case "exit", "quit":
			break LOOP
		case "help":
			usage()
		default:
			cmd, ok := commands[args[0]]
			if ok {
				numargs := len(args) - 1
				if numargs < cmd.args {
					fmt.Printf("Not enough arguments given to %v. Needs at least %v, given %v\n", args[0], cmd.args, numargs)
					fmt.Printf("Usage: %v\n", cmd.usage)
				} else {
					err := cmd.f(shell, args)
					if err != nil {
						fmt.Printf("Error while running %v: %v\n", args[0], err)
					}
				}
			} else {
				fmt.Println("Unrecognized command. Type \"help\" to see available commands.")
			}
		}
	}
	shell.done <- true
}

func optionBool(opt string) (bool, error) {
	switch strings.ToLower(opt) {
	case "on", "true":
		return true, nil
	case "off", "false":
		return false, nil
	default:
		return false, fmt.Errorf("Unknown state %s. Expect on or off. ", opt)
	}
}
