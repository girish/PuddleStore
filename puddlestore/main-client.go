// +build client

package main

import (
	"./puddlestore"
	"flag"
	"fmt"
)

func main() {
	var addr string
	var debug bool

	addrHelpString := "An online node of the Raft cluster to connect to. If left blank, does not attempt to connect to another node."
	flag.StringVar(&addr, "connect", "", addrHelpString)
	flag.StringVar(&addr, "c", "", addrHelpString)

	flag.BoolVar(&debug, "debug", false, "Turn on debug message printing.")
	flag.BoolVar(&debug, "d", false, "Turn on debug message printing. (shorthand)")

	flag.Parse()

	puddlestore.SetDebug(debug)

	// config := raft.DefaultConfig()

	var remoteAddr puddlestore.PuddleAddr
	if addr == "" {
		fmt.Printf("You must specify an address for the client to connect to!\n")
		return
	} else {
		remoteAddr = puddlestore.PuddleAddr{addr}
	}

	c, err := puddlestore.CreateClient(remoteAddr)

	if err != nil {
		fmt.Printf("Error starting client: %v\n", err)
		return
	}

	clientCommands := map[string]command{
		// "debug": command{toggleDebug, "debug <on|off>", "Turn debug on or off. On by default", 1},
		"ls":    command{ls, "ls", "list directory contents", 0},
		"cd":    command{cd, "cd <path>", "change to given directory", 1},
		"mkdir": command{mkdir, "mkdir <path>", "make directory", 1},
	}

	// Kick off CLI, await exit
	var shell Shell
	shell.done = make(chan bool)
	shell.c = c
	go shell.interact(clientCommands)

	for !(<-shell.done) {
	}

}
