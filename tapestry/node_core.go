/*
 *  Brown University, CS138, Spring 2020
 *
 *  Purpose: Defines functions to publish and lookup objects in a Tapestry mesh
 */

package tapestry

// __BEGIN_TA__
import (
	"fmt"
	"time"
	// xtr "github.com/brown-csci1380/tracing-framework-go/xtrace/client"
)

// __END_TA__
/* __BEGIN_STUDENT__
import (
	"fmt"
	// Uncomment for xtrace
	// xtr "github.com/brown-csci1380/tracing-framework-go/xtrace/client"
)
__END_STUDENT__ */

// Store a blob on the local node and publish the key to the tapestry.
func (local *Node) Store(key string, value []byte) (err error) {
	done, err := local.Publish(key)
	if err != nil {
		return err
	}
	local.blobstore.Put(key, value, done)
	return nil
}

// Get looks up a key in the tapestry then fetch the corresponding blob from the
// remote blob store.
func (local *Node) Get(key string) ([]byte, error) {
	// Lookup the key
	replicas, err := local.Lookup(key)
	if err != nil {
		return nil, err
	}
	if len(replicas) == 0 {
		return nil, fmt.Errorf("No replicas returned for key %v", key)
	}

	// Contact replicas
	var errs []error
	for _, replica := range replicas {
		blob, err := replica.BlobStoreFetchRPC(key)
		if err != nil {
			errs = append(errs, err)
		}
		if blob != nil {
			return *blob, nil
		}
	}

	return nil, fmt.Errorf("Error contacting replicas, %v: %v", replicas, errs)
}

// Remove the blob from the local blob store and stop advertising
func (local *Node) Remove(key string) bool {
	return local.blobstore.Delete(key)
}

// Publishes the key in tapestry.
//
// - Start periodically publishing the key. At each publishing:
// 		- Find the root node for the key
// 		- Register the local node on the root
// 		- if anything failed, retry; until RETRIES has been reached.
// - Return a channel for cancelling the publish
// 		- if receiving from the channel, stop republishing
func (local *Node) Publish(key string) (cancel chan bool, err error) {
	// __BEGIN_TA__
	// xtr.NewTask("publish")

	publish := func(key string) error {
		Debug.Printf("Publishing %v\n", key)

		failures := 0
		for failures < RETRIES {
			// Route to the root node
			root, _, err := local.FindRoot(Hash(key), 0)
			if err != nil {
				failures++
				continue
			}

			// Register our local node on the root
			isRoot, err := root.RegisterRPC(key, local.node)
			if err != nil {
				// xtr.AddTags("failure")
				// Trace.Printf("failed to publish, bad node: %v\n", root)
				local.RemoveBadNodes([]RemoteNode{root})
				failures++
			} else if !isRoot {
				Trace.Printf("failed to publish to %v, not the root node", root)
				failures++
			} else {
				Trace.Printf("Successfully published %v on %v", key, root)
				return nil
			}
		}

		// xtr.AddTags("failure")
		// Trace.Printf("failed to publish %v (%v) due to %v/%v failures", key, Hash(key), failures, RETRIES)
		return fmt.Errorf("Unable to publish %v (%v) due to %v/%v failures", key, Hash(key), failures, RETRIES)
	}

	// Publish the key immediately
	err = publish(key)
	// (Optional) quits if the first attempt fails. Store is rejected.
	if err != nil {
		return
	}

	// Create the cancel channel
	cancel = make(chan bool)

	// Periodically republish the key
	go func() {
		ticker := time.NewTicker(REPUBLISH)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				{
					// xtr.NewTask("republish")
					Trace.Printf("republishing %v", key)
					err := publish(key)
					if err != nil {
						Error.Print(err)
					}
				}
			case <-cancel:
				{
					Trace.Printf("Stopping advertisement of %v", key)
					fmt.Printf("Stopping advertisement for %v\n", key)
					return
				}
			}
		}
	}()
	// __END_TA__
	// __BEGIN_STUDENT__
	// TODO: students should implement this
	// __END_STUDENT__
	return
}

// Lookup look up the Tapestry nodes that are storing the blob for the specified key.
//
// - Find the root node for the key
// - Fetch the replicas (nodes storing the blob) from the root's location map
// - Attempt up to RETRIES times
func (local *Node) Lookup(key string) (nodes []RemoteNode, err error) {
	// __BEGIN_TA__
	// xtr.NewTask("lookup")
	Trace.Printf("Looking up %v", key)

	// Function to look up a key
	lookup := func(key string) ([]RemoteNode, error) {
		// Look up the root node
		node, _, err := local.FindRoot(Hash(key), 0)
		//node, err := local.findRootOnRemoteNode(local.node, Hash(key))
		if err != nil {
			return nil, err
		}

		// Get the replicas from the root's location map
		isRoot, nodes, err := node.FetchRPC(key)
		if err != nil {
			return nil, err
		} else if !isRoot {
			return nil, fmt.Errorf("Root node did not believe it was the root node")
		} else {
			return nodes, nil
		}
	}

	// Attempt up to RETRIES many times
	errs := make([]error, 0, RETRIES)
	for len(errs) < RETRIES {
		Debug.Printf("Looking up %v, attempt=%v\n", key, len(errs)+1)
		results, lookup_err := lookup(key)
		if lookup_err != nil {
			Error.Println(lookup_err)
			errs = append(errs, lookup_err)
		} else {
			return results, nil
		}
	}

	err = fmt.Errorf("%v failures looking up %v: %v", RETRIES, key, errs)

	// __END_TA__
	// __BEGIN_STUDENT__
	// TODO: students should implement this
	// __END_STUDENT__
	return
}

// FindRoot returns the root for id by recursive RPC calls on the next hop found in our routing table
// 		- find the next hop from our routing table
// 		- call FindRoot on nextHop
// 		- if failed, add nextHop to toRemove, remove them from local routing table, retry
func (local *Node) FindRoot(id ID, level int32) (root RemoteNode, toRemove *NodeSet, err error) {
	// __BEGIN_TA__
	toRemove = NewNodeSet()
	for {
		nextHop := local.table.FindNextHop(id, level)
		if nextHop == local.node {
			return nextHop, toRemove, nil
		}

		// recursively call FindRoot on nextHop next hop
		var newRemove *NodeSet
		root, newRemove, err = nextHop.FindRootRPC(id, level+1)
		toRemove.AddAll(newRemove.Nodes())
		local.RemoveBadNodes(toRemove.Nodes())

		if err != nil {
			toRemove.Add(nextHop)
			// immediately remove stale next hop from our routing table so it won't be found next time
			local.table.Remove(nextHop)
		} else {
			break
		}
	}

	return root, toRemove, nil
	// __END_TA__

	// TODO: students should implement this
	return
}

// The replica that stores some data with key is registering themselves to us as an advertiser of the key.
// - Check that we are the root node for the key, set `isRoot`
// - Add the node to the location map (local.locationsByKey.Register)
// 		- local.locationsByKey.Register kicks off a timer to remove the node if it's not advertised again
// 		  after TIMEOUT
func (local *Node) Register(key string, replica RemoteNode) (isRoot bool) {
	// __BEGIN_TA__
	node, _, _ := local.FindRoot(Hash(key), 0)
	if node == local.node {
		isRoot = true
		if local.locationsByKey.Register(key, replica, TIMEOUT) {
			Debug.Printf("Register %v:%v (%v)\n", key, replica, Hash(key))
		}
	} else {
		return false
	}
	// __END_TA__
	// __BEGIN_STUDENT__
	// TODO: students should implement this
	// __END_STUDENT__
	return
}

// Fetch checks that we are the root node for the requested key and
// return all nodes that are registered in the local location map for this key
func (local *Node) Fetch(key string) (isRoot bool, replicas []RemoteNode) {
	// __BEGIN_TA__
	node, _, _ := local.FindRoot(Hash(key), 0)
	if node == local.node {
		isRoot = true
		replicas = local.locationsByKey.Get(key)
		Debug.Printf("Lookup %v:%v (%v)\n", key, replicas, Hash(key))
	} else {
		isRoot = false
	}
	// __END_TA__
	// __BEGIN_STUDENT__
	// TODO: students should implement this
	// __END_STUDENT__
	return
}

// Transfer registers all of the provided objects in the local location map. (local.locationsByKey.RegisterAll)
// If appropriate, add the from node to our local routing table
func (local *Node) Transfer(from RemoteNode, replicaMap map[string][]RemoteNode) (err error) {
	// __BEGIN_TA__
	if len(replicaMap) > 0 {
		Debug.Printf("Registering objects from %v: %v\n", from, replicaMap)
		local.locationsByKey.RegisterAll(replicaMap, TIMEOUT)
	}
	local.addRoute(from)
	// __END_TA__
	// __BEGIN_STUDENT__
	// TODO: students should implement this
	// __END_STUDENT__
	return nil
}

// calls FindRoot on a remote node with given ID
func (local *Node) findRootOnRemoteNode(start RemoteNode, id ID) (RemoteNode, error) {
	//__BEGIN_TA__

	// Keep track of faulty nodes along the way
	root, _, err := start.FindRootRPC(id, 0)
	if err != nil {
		return RemoteNode{}, fmt.Errorf("unable to get root for %v, all nodes traversed were bad, starting from %v", id, start)
	}
	return root, err
	// __END_TA__
	// TODO: students should implement this
}
