package main

import (
	"./tapestry"
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

	tapestry.SetDebug(debug)

	// Print
	switch {
	case port != 0 && addr != "":
		{
			tapestry.Out.Printf("Starting a node on port %v and connecting to %v\n", port, addr)
		}
	case port != 0:
		{
			tapestry.Out.Printf("Starting a standalone node on port %v\n", port)
		}
	case addr != "":
		{
			tapestry.Out.Printf("Starting a node on a random port and connecting to %v\n", addr)
		}
	default:
		{
			tapestry.Out.Printf("Starting a standalone node on a random port\n")
		}
	}

	t, err := tapestry.Start(port, addr)

	if err != nil {
		fmt.Printf("Error starting tapestry node: %v\n", err)
		return
	}

	tapestry.Out.Printf("Successfully started: %v\n", t)

	// Kick off CLI, await exit
	done := make(chan bool)
	go CLI(t, done)

	for !(<-done) {
	}

	tapestry.Out.Println("Closing tapestry")

	t.Leave()
}
