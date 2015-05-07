package main

import (
	"./puddlestore"
	"fmt"
)

func toggleDebug(shell *Shell, args []string) error {
	state, err := optionBool(args[1])
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	puddlestore.SetDebug(state)
	return nil
}

/*
func showState(shell *Shell, args []string) error {
	shell.p.ShowState()
	return nil
}

func showLog(shell *Shell, args []string) error {
	shell.r.PrintLogCache()
	return nil
}

func enable(shell *Shell, args []string) error {
	shell.r.Testing.PauseWorld(false)
	return nil
}

func disable(shell *Shell, args []string) error {
	shell.r.Testing.PauseWorld(true)
	return nil
}

func sendrecv(shell *Shell, args []string) error {
	state, err := optionBool(args[2])
	if err != nil {
		return err
	}
	node := findNode(shell.r.GetOtherNodes(), args[1])
	if node == nil {
		return fmt.Errorf("Given string doesn't match any nodes' ID or address")
	}
	if args[0] == "send" {
		shell.r.Testing.RegisterPolicy(*shell.r.GetLocalAddr(), *node, state)
	} else {
		shell.r.Testing.RegisterPolicy(*node, *shell.r.GetLocalAddr(), state)
	}
	return nil
}

func findNode(nodes []raft.NodeAddr, id string) *raft.NodeAddr {
	for _, node := range nodes {
		if (node.Id == id) ||
			(node.Addr == id) {
			return &node
		}
	}
	return nil
}*/
