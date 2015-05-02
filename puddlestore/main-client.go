// +build client

package main

import (
	"./raft"
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

	raft.SetDebug(debug)

	config := raft.DefaultConfig()

	var remoteAddr raft.NodeAddr
	if addr == "" {
		fmt.Printf("You must specify an address for the client to connect to!\n")
		return
	} else {
		remoteAddr = raft.NodeAddr{raft.AddrToId(addr, config.NodeIdSize), addr}
	}

	c, err := raft.CreateClient(remoteAddr)

	if err != nil {
		fmt.Printf("Error starting client: %v\n", err)
		return
	}

	clientCommands := map[string]command{
		"debug": command{toggleDebug, "debug <on|off>", "Turn debug on or off. On by default", 1},
		"init":  command{clientInit, "init <value>", "Instruct the state machine to pick an initial value for hashing", 1},
		"hash":  command{clientHash, "hash", "Instruct the state machine to perform another round of hashing", 0},
	}

	// Kick off CLI, await exit
	var shell Shell
	shell.done = make(chan bool)
	shell.c = c
	go shell.interact(clientCommands)

	for !(<-shell.done) {
	}

}
