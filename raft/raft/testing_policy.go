package raft

import (
	"errors"
	"fmt"
)

/* */
type TestingPolicy struct {
	pauseWorld bool
	rpcPolicy  map[string]bool
}

func NewTesting() *TestingPolicy {
	var tp TestingPolicy
	tp.rpcPolicy = make(map[string]bool)
	return &tp
}

var ErrorTestingPolicyDenied = errors.New("the testing policy has forbid this communication")

func getCommId(a, b NodeAddr) string {
	if a.Id < b.Id {
		return fmt.Sprintf("%v_%v", a.Id, b.Id)
	} else {
		return fmt.Sprintf("%v_%v", b.Id, a.Id)
	}
}

/*                                                                     */
/* Check our testing policy to see if we are allowed to send or        */
/* receive messages with this node.                                    */
/*                                                                     */
func (tp *TestingPolicy) IsDenied(a, b NodeAddr) bool {
	if tp.pauseWorld {
		return true
	}
	commStr := getCommId(a, b)
	allowed, exists := tp.rpcPolicy[commStr]
	return exists && !allowed
}

func (tp *TestingPolicy) RegisterPolicy(a, b NodeAddr, allowed bool) {
	commStr := getCommId(a, b)
	tp.rpcPolicy[commStr] = allowed
}

func (tp *TestingPolicy) PauseWorld(on bool) {
	tp.pauseWorld = on
}
