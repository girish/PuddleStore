// +build !client

package main

import (
	"./puddlestore"
	"flag"
	"fmt"
)

func main() {
	var port int
	var addr string
	var debug bool

	flag.IntVar(&port, "port", 0, "The server port to bind to. Defaults to a random port.")
	flag.IntVar(&port, "p", 0, "The server port to bind to. Defaults to a random port. (shorthand)")

	flag.StringVar(&addr, "connect", "", "An existing node to connect to. If left blank, does not attempt to connect to another node.")
	flag.StringVar(&addr, "c", "", "An existing node to connect to. If left blank, does not attempt to connect to another node.  (shorthand)")

	flag.BoolVar(&debug, "debug", false, "Turn on debug message printing.")
	flag.BoolVar(&debug, "d", false, "Turn on debug message printing. (shorthand)")

	flag.Parse()

	puddlestore.SetDebug(debug)

	// config := raft.DefaultConfig()

	// var remoteAddr *puddlestore.PuddleAddr
	if addr != "" {
		// remoteAddr = &puddlestore.PuddleAddr{addr}
	}

	r, err := puddlestore.Start()

	if err != nil {
		fmt.Printf("Error starting puddle node: %v\n", err)
		return
	}

	puddlestore.Out.Printf("Successfully started: %v\n", r)

	nodeCommands := map[string]command{
		"debug": command{toggleDebug, "debug <on|off>", "Turn debug on or off. On by default", 1},
		/*
			"state":   command{showState, "state", "Print out the current local and cluster state", 0},
			"log":     command{showLog, "log", "Print out the local log cache", 0},
			"disable": command{disable, "disable", "Prevent this node from communicating with the cluster", 0},
			"enable":  command{enable, "enable", "Allow this node to communicate with the cluster", 0},
			"send":    command{sendrecv, "send <addr> <on|off>", "Prevent this node from sending to a given address", 2},
			"recv":    command{sendrecv, "recv <addr> <on|off>", "Prevent this node from receiving from a given address", 2},
		*/
	}

	// Kick off CLI, await exit
	var shell Shell
	shell.done = make(chan bool)
	shell.p = r
	go shell.interact(nodeCommands)

	for !(<-shell.done) {
	}

	puddlestore.Out.Println("Closing local raft node")

	// r.GracefulExit()

	puddlestore.Out.Println("Bye!")
}
