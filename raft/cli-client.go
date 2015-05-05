package main

import (
	"./raft"
)

func clientInit(shell *Shell, args []string) error {
	return shell.c.SendRequest(raft.HASH_CHAIN_INIT, []byte(args[1]))
}

func clientHash(shell *Shell, args []string) error {
	return shell.c.SendRequest(raft.HASH_CHAIN_ADD, []byte{})
}

// This function puts something in the map
func clientSet(shell *Shell, args []string) error {
	return shell.c.SendRequest(raft.SET, []byte(args[1]+":"+args[2]))
}

//This function gets something in the map
func clientGet(shell *Shell, args []string) error {
	return shell.c.SendRequest(raft.GET, []byte(args[1]))
}

//This function removes somehting form the log
func clientRemove(shell *Shell, args []string) error {
	return shell.c.SendRequest(raft.REMOVE, []byte(args[1]))
}
