package tapestry

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"
	//t "tapestry/tapestry"
)

func hasnode(slice []RemoteNode, item RemoteNode) bool {
	set := make(map[RemoteNode]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}

func hasneighbor(slice []RemoteNode, item RemoteNode) bool {
	return hasnode(slice, item)
}

func hasBackpointer(t1 *Node, t2 *Node) bool {
	return hasBackpointerNode(t1, t2.node)
}
func hasBackpointerNode(t1 *Node, n2 RemoteNode) bool {
	id1 := t1.node.ID
	id2 := n2.ID
	if id1 == id2 {
		return false
	} else {
		spl := 0
		for ; spl < DIGITS; spl++ {
			if id1[spl] != id2[spl] {
				break
			}
		}

		nodeset := t1.backpointers.sets[spl]
		if nodeset == nil {
			return false
		}
		return nodeset.Contains(n2)
	}

}

func hasRoutingTableEntry(t1 *Node, t2 *Node) bool {
	return hasRoutingTableNode(t1, t2.node)
}

func hasRoutingTableNode(t1 *Node, n2 RemoteNode) bool {
	return t1.table.TAContains(n2)
}

/*
   Parse an ID from String
*/
func PartialID(stringID string) (ID, error) {
	var id ID

	for i := 0; i < DIGITS && i < len(stringID); i++ {
		d, err := strconv.ParseInt(stringID[i:i+1], 16, 0)
		if err != nil {
			return id, err
		}
		id[i] = Digit(d)
	}
	for i := len(stringID); i < DIGITS; i++ {
		id[i] = Digit(0)
	}

	return id, nil
}

/*
   Parse an ID from String
*/
func MakeID(stringID string) ID {
	var id ID

	for i := 0; i < DIGITS && i < len(stringID); i++ {
		d, err := strconv.ParseInt(stringID[i:i+1], 16, 0)
		if err != nil {
			return id
		}
		id[i] = Digit(d)
	}
	for i := len(stringID); i < DIGITS; i++ {
		id[i] = Digit(0)
	}

	return id
}

func TapestryPause() {
	time.Sleep(200 * time.Millisecond)
}

var tapestriesByAddress map[string]*Node = make(map[string]*Node)
var tapestryMapMutex *sync.Mutex = &sync.Mutex{}

func registerCachedTapestry(tapestry ...*Node) {
	tapestryMapMutex.Lock()
	defer tapestryMapMutex.Unlock()
	for _, t := range tapestry {
		tapestriesByAddress[t.node.Address] = t
	}
}

func unregisterCachedTapestry(tapestry ...*Node) {
	tapestryMapMutex.Lock()
	defer tapestryMapMutex.Unlock()
	for _, t := range tapestry {
		delete(tapestriesByAddress, t.node.Address)
	}
}

func MakeOne(ida string) (t1 *Node, err error) {
	t1, err = start(MakeID(ida), 0, "")
	if err != nil {
		return nil, err
	}
	registerCachedTapestry(t1)
	return
}
func AddOne(ida string, addr string, tap []*Node) (t1 *Node, tapNew []*Node, err error) {
	t1, err = start(MakeID(ida), 0, addr)
	if err != nil {
		return nil, tap, err
	}
	registerCachedTapestry(t1)
	tapNew = append(tap, t1)
	time.Sleep(1000 * time.Millisecond) //Wait for availability
	return
}
func MakeTwo(ida, idb string) (t1 *Node, t2 *Node, err error) {
	t1, err = start(MakeID(ida), 0, "")
	if err != nil {
		return nil, nil, err
	}
	t2, err = start(MakeID(idb), 0, t1.node.Address)
	if err != nil {
		return nil, nil, err
	}
	registerCachedTapestry(t1, t2)
	return
}

func MakeThree(ida, idb, idc string) (t1 *Node, t2 *Node, t3 *Node, err error) {
	t1, err = start(MakeID(ida), 0, "")
	if err != nil {
		return nil, nil, nil, err
	}
	t2, err = start(MakeID(idb), 0, t1.node.Address)
	if err != nil {
		return nil, nil, nil, err
	}
	t3, err = start(MakeID(idc), 0, t1.node.Address)
	if err != nil {
		return nil, nil, nil, err
	}
	registerCachedTapestry(t1, t2, t3)
	return
}

func MakeTapestries(connectThem bool, ids ...string) ([]*Node, error) {
	tapestries := make([]*Node, 0, len(ids))
	for i := 0; i < len(ids); i++ {
		connectTo := ""
		if i > 0 && connectThem {
			connectTo = tapestries[0].node.Address
		}
		t, err := start(MakeID(ids[i]), 0, connectTo)
		if err != nil {
			return tapestries, err
		}
		registerCachedTapestry(t)
		tapestries = append(tapestries, t)
		time.Sleep(10 * time.Millisecond)
	}
	return tapestries, nil
}

func KillTapestries(ts ...*Node) {
	fmt.Println("killing")
	unregisterCachedTapestry(ts...)
	for _, t := range ts {
		t.Kill()
	}
	fmt.Println("finished killing")
}

/*
	Returns true if the node exists anywhere the routing table
*/
func (t *RoutingTable) TAContains(node RemoteNode) (contains bool) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	for i := 0; i < DIGITS; i++ {
		for j := 0; j < BASE; j++ {
			slot := t.rows[i][j]
			if slot != nil {
				for k := 0; k < len(slot); k++ {
					if slot[k] == node {
						return true
					}
				}
			}
		}
	}
	return false
}

func MakeRandomTapestries(seed int64, count int) ([]*Node, error) {
	r := rand.New(rand.NewSource(seed))

	ts := make([]*Node, 0, count)

	for i := 0; i < count; i++ {
		connectTo := ""
		if i > 0 {
			connectTo = ts[0].node.Address
		}
		t, err := start(IntToID(r.Int()), 0, connectTo)
		if err != nil {
			return ts, err
		}
		registerCachedTapestry(t)
		ts = append(ts, t)
		time.Sleep(10 * time.Millisecond)
	}

	return ts, nil
}

func MakeMoreRandomTapestries(seed int64, count int, ts []*Node) ([]*Node, error) {
	r := rand.New(rand.NewSource(seed + int64(len(ts))))

	for i := 0; i < count; i++ {
		connectTo := ts[0].node.Address
		t, err := start(IntToID(r.Int()), 0, connectTo)
		if err != nil {
			return ts, err
		}
		registerCachedTapestry(t)
		ts = append(ts, t)
		time.Sleep(10 * time.Millisecond)
	}
	return ts, nil
}

func Candidates(seed int64, count int) []ID {
	r := rand.New(rand.NewSource(seed + 999))

	ids := make([]ID, 0, count)

	for i := 0; i < count; i++ {
		ids = append(ids, IntToID(r.Int()))
	}

	return ids
}

func IntToID(x int) ID {
	var id ID
	for i := range id {
		id[i] = Digit(x % BASE)
		x = x / BASE
	}
	return id
}

type PublishSpec struct {
	store  *Node
	lookup *Node
	key    string
}

func GenerateData(seed int64, count int, ts []*Node) []PublishSpec {
	specs := make([]PublishSpec, 0, count)
	r := rand.New(rand.NewSource(seed + 499999))

	for i := 0; i < count; i++ {
		storeI := r.Intn(len(ts))
		lookupI := r.Intn(len(ts))

		store := ts[storeI]
		lookup := ts[lookupI]
		key := fmt.Sprintf("%v-%v-%v", i, storeI, lookupI)

		specs = append(specs, PublishSpec{store, lookup, key})
	}

	return specs
}

func GenerateKeys(seed int64, count int) []string {
	keys := make([]string, 0, count+9999)

	for _, i := range GenerateInts(seed, count) {
		keys = append(keys, fmt.Sprintf("%v", i))
	}

	return keys
}

func GenerateInts(seed int64, count int) []int {
	r := rand.New(rand.NewSource(seed + 99999))

	ints := make([]int, 0, count)

	for i := 0; i < count; i++ {
		ints = append(ints, r.Int())
	}

	return ints
}

// consistent with student util
func ContainsNode(l []RemoteNode, n RemoteNode) bool {
	for _, node := range l {
		if node == n {
			return true
		}
	}
	return false
}
