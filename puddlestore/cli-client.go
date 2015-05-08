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
	var output string
	var err error
	if len(args) > 1 {
		output, err = shell.c.Ls(args[1])
	} else {
		output, err = shell.c.Ls("")
	}

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(output)
	}
	return nil
}

func cd(shell *Shell, args []string) error {
	var err error
	if len(args) > 1 {
		err = shell.c.Cd(args[1])
	} else {
		err = shell.c.Cd("")
	}
	if err != nil {
		fmt.Println(err)
	}
	return nil
}

func mkdir(shell *Shell, args []string) error {
	err := shell.c.Mkdir(args[1])
	if err != nil {
		fmt.Println(err)
	}
	return nil
}

<<<<<<< HEAD
func mkfile(shell *Shell, args []string) error {
	fmt.Println("Running mkfile with", args[1])
	err := shell.c.Mkfile(args[1])
=======
func rmdir(shell *Shell, args []string) error {
	err := shell.c.Rmdir(args[1])
>>>>>>> 0f76b02... Implements rmdir
	if err != nil {
		fmt.Println(err)
	}
	return nil
}
