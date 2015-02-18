package tapestry

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var Debug *log.Logger
var Out *log.Logger
var Error *log.Logger

// Initialize the loggers
func init() {
	Debug = log.New(ioutil.Discard, "", log.Ltime|log.Lshortfile)
	Out = log.New(os.Stdout, "", log.Ltime|log.Lshortfile)
	Error = log.New(os.Stdout, "ERROR: ", log.Ltime|log.Lshortfile)
}

// Turn debug on or off
func SetDebug(enabled bool) {
	if enabled {
		Debug = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		Debug = log.New(ioutil.Discard, "", log.Ldate|log.Ltime|log.Lshortfile)
	}
}

// Prints a routing table
func (tapestry *Tapestry) PrintRoutingTable() {
	table := tapestry.local.table
	fmt.Printf("RoutingTable for node %v\n", table.local)
	id := table.local.Id.String()
	for i, row := range table.rows {
		for j, slot := range row {
			for _, node := range *slot {
				fmt.Printf(" %v%v  %v: %v %v\n", id[:i], strings.Repeat(" ", DIGITS-i+1), Digit(j), node.Address, node.Id.String())
			}
		}
	}
}

// Prints the object store
func (tapestry *Tapestry) PrintObjectStore() {
	fmt.Printf("ObjectStore for node %v\n", tapestry.local.node)
	for key, values := range tapestry.local.store.data {
		fmt.Printf(" %v: %v\n", key, slice(values))
	}
}

// Prints the backpointers
func (tapestry *Tapestry) PrintBackpointers() {
	bp := tapestry.local.backpointers
	fmt.Printf("Backpointers for node %v\n", tapestry.local.node)
	for i, set := range bp.sets {
		for _, node := range set.Nodes() {
			fmt.Printf(" %v %v: %v\n", i, node.Address, node.Id.String())
		}
	}
}

// Prints the blobstore
func (tapestry *Tapestry) PrintBlobStore() {
	for k, _ := range tapestry.blobstore.blobs {
		fmt.Println(k)
	}
}
