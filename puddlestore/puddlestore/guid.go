package puddlestore

/*
	File from puddlestore in charge of dealing with raft.
*/

import (
	"../../raft/raft"
	"fmt"
)

func (puddle *PuddleNode) setRaftVguid(aguid Aguid, vguid Vguid, id uint64) error {
	// Get the raft client struct
	c, ok := puddle.clients[id]
	if !ok {
		panic("Attempted to get client from id, but not found.")
	}

	data := fmt.Sprintf("%v:%v", aguid, vguid)

	res, err := c.SendRequestWithResponse(raft.SET, []byte(data))
	if err != nil {
		return err
	}
	if res.Status != raft.OK {
		return fmt.Errorf("Could not get response from raft.")
	}

	return nil
}

func (puddle *PuddleNode) getRaftVguid(aguid Aguid, id uint64) (Vguid, error) {
	// Get the raft client struct
	c, ok := puddle.clients[id]
	if !ok {
		panic("Attempted to get client from id, but not found.")
	}

	res, err := c.SendRequestWithResponse(raft.GET, []byte(aguid))
	if err != nil {
		return "", err
	}
	if res.Status != raft.OK {
		return "", fmt.Errorf("Could not get response from raft.")
	}

	return Vguid(res.Response), nil
}

func (puddle *PuddleNode) removeRaftVguid(aguid Aguid, id uint64) error {
	// Get the raft client struct
	c, ok := puddle.clients[id]
	if !ok {
		panic("Attempted to get client from id, but not found.")
	}

	res, err := c.SendRequestWithResponse(raft.REMOVE, []byte(aguid))
	if err != nil {
		return err
	}
	if res.Status != raft.OK {
		return fmt.Errorf("Could not get response from raft.")
	}

	return nil
}
