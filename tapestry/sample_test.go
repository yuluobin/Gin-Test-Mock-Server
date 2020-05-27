package tapestry

import (
	"bytes"
	"fmt"
	"testing"
)

func TestSampleTapestrySetup(t *testing.T) {
	tap, _ := MakeTapestries(true, "1", "3", "5", "7") //Make a tapestry with these ids
	fmt.Printf("length of tap %d\n", len(tap))
	KillTapestries(tap[1], tap[2])                //Kill off two of them.
	next, _, _ := tap[0].FindRoot(MakeID("2"), 0) //After killing 3 and 5, this should route to 7
	if next != tap[3].node {
		t.Errorf("Failed to kill successfully")
	}

}

func TestSampleTapestrySearch(t *testing.T) {
	tap, _ := MakeTapestries(true, "100", "456", "1234") //make a sample tap
	tap[1].Store("look at this lad", []byte("an absolute unit"))
	result, _ := tap[0].Get("look at this lad")           //Store a KV pair and try to fetch it
	if !bytes.Equal(result, []byte("an absolute unit")) { //Ensure we correctly get our KV
		t.Errorf("Get failed")
	}
}

func TestSampleTapestryAddNodes(t *testing.T) {
	tap, _ := MakeTapestries(true, "1", "5", "9")
	node8, tap, _ := AddOne("8", tap[0].node.Address, tap) //Add some tap nodes after the initial construction
	_, tap, _ = AddOne("12", tap[0].node.Address, tap)

	next, _, _ := tap[1].FindRoot(MakeID("7"), 0)
	if node8.node != next {
		t.Errorf("Addition of node failed")
	}
}
