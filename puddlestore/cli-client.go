package main

import (
	"fmt"
)

/*
func clientInit(shell *Shell, args []string) error {
	return shell.c.SendRequest(raft.HASH_CHAIN_INIT, []byte(args[1]))
}

func clientHash(shell *Shell, args []string) error {
	return shell.c.SendRequest(raft.HASH_CHAIN_ADD, []byte{})
}
*/

func ls(shell *Shell, args []string) error {
	output, err := shell.c.Ls()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(output)
	}
	return nil
}

func cd(shell *Shell, args []string) error {
	return nil
}

func mkdir(shell *Shell, args []string) error {
	fmt.Println("Runnink mkdir with", args[1])
	err := shell.c.Mkdir(args[1])
	if err != nil {
		fmt.Println(err)
	}
	return nil
}
