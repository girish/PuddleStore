/*                                                                           */
/*  Brown University, CS138, Spring 2015                                     */
/*                                                                           */
/*  Purpose: API and interal functions to interact with the Key-Value store  */
/*           that the Chord ring is providing.                               */
/*                                                                           */

package chord

import (
	"fmt"
)

/*                             */
/* External API Into Datastore */
/*                             */

/* Get a value in the datastore, provided an abitrary node in the ring */
func Get(node *Node, key string) (string, error) {
	remNode, err := node.locate(key)
	if err != nil {
		return "", err
	}
	fmt.Printf("Executing get one %v", remNode.Id)
	return Get_RPC(remNode, key)
}

/* Put a key/value in the datastore, provided an abitrary node in the ring */
func Put(node *Node, key string, value string) error {
	remNode, err := node.locate(key)
	if err != nil {
		return err
	}
	return Put_RPC(remNode, key, value)
}

/* Internal helper method to find the appropriate node in the ring */
func (node *Node) locate(key string) (*RemoteNode, error) {
	id := HashKey(key)
	return node.findSuccessor(id)
}

/* When we discover a new predecessor we may need to transfer some keys to it */
/*Oh I think I get it, this one is to send
This was eliminated by the TAs because of its redundancy */
func (node *Node) obtainNewKeys() error {
	//lock the local db and get the keys
	(&node.dsLock).Lock()
	for key, val := range node.dataStore {
		keyByte := HashKey(key)
		if !BetweenRightIncl(keyByte, node.Predecessor.Id, node.Id) {
			//means we send it to the predecessor
			err := Put_RPC(node.Predecessor, key, val)
			if err != nil {
				//TODO handle error, particularly decide what to do with the ones not transfered
				(&node.dsLock).Unlock()
				return err
			}
			//then we delete it locally
			delete(node.dataStore, key)
		}
	}
	//unlock the db
	(&node.dsLock).Unlock()
	return nil
}

/*                                                         */
/* RPCs to assist with interfacing with the datastore ring */
/*                                                         */

/* RPC */
//This is in response to an RPC made onto us.
func (node *Node) GetLocal(req *KeyValueReq, reply *KeyValueReply) error {
	if err := validateRpc(node, req.NodeId); err != nil {
		return err
	}
	fmt.Printf("%p", node.dsLock)
	(&node.dsLock).RLock()
	fmt.Printf("Executing get local 2")
	key := req.Key
	fmt.Printf("Executing get local 3")
	val := node.dataStore[key]
	fmt.Printf("Executing get local 4")
	reply.Key = key
	fmt.Printf("Executing get local 5")
	reply.Value = val
	fmt.Printf("Executing get local 6")
	(&node.dsLock).RUnlock()
	fmt.Printf("Executing get local 7")
	return nil
}

/* RPC */
func (node *Node) PutLocal(req *KeyValueReq, reply *KeyValueReply) error {
	if err := validateRpc(node, req.NodeId); err != nil {
		return err
	}
	fmt.Printf("%p", node.dsLock)
	(&node.dsLock).Lock()
	key := req.Key
	val := req.Value
	node.dataStore[key] = val
	reply.Key = key
	reply.Value = val
	(&node.dsLock).Unlock()
	return nil
}

/* RPC OLD
This function call is called on us as the successor. This is suppose to trigger us to transfer the relevant
keys back to node*/
/* Comment from the TAs: Find locally stored keys that are between (predId : fromId],
any of these nodes should be moved to fromId */

/* RPC */
/* Find locally stored keys that are between (predId : fromId], any of
   these nodes should be moved to fromId */
func (node *Node) TransferKeys(req *TransferReq, reply *RpcOkay) error {
	if err := validateRpc(node, req.NodeId); err != nil {
		return err
	}
	(&node.dsLock).Lock()
	//fmt.Println("ok va")
	for key, val := range node.dataStore {
		keyByte := HashKey(key)
		pred := req.PredId
		if pred == nil {
			pred = node.Id
		}
		if BetweenRightIncl(keyByte, pred, req.FromId) {
			//means we send it to the requester, because it belongs to them
			err := Put_RPC(node.Predecessor, key, val)
			if err != nil {
				//TODO handle error, particularly decide what to do with the ones not transfered
				(&node.dsLock).Unlock()
				reply.Ok = false
				return err
			}
			//then we delete it locally
			delete(node.dataStore, key)
		}
	}
	//unlock the db
	//node.dsLock.Unlock()
	//if err != nil {
	//	reply.Ok = false
	//	return err
	//}
	reply.Ok = true
	return nil
}

/* Print the contents of a node's data store */
func PrintDataStore(node *Node) {
	fmt.Printf("Node-%v datastore: %v\n", HashStr(node.Id), node.dataStore)
}
