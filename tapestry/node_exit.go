/*
 *  Brown University, CS138, Spring 2020
 *
 *  Purpose: Defines functions for a node leaving the Tapestry mesh, and
 *  transferring its stored locations to a new node.
 */

package tapestry

// __BEGIN_TA__
// import (
// 	xtr "github.com/brown-csci1380/tracing-framework-go/xtrace/client"
// )

// __END_TA__
/* __BEGIN_STUDENT__
import (
	// Uncomment for xtrace
	// xtr "github.com/brown-csci1380/tracing-framework-go/xtrace/client"
)
__END_STUDENT__ */

// Kill this node without gracefully leaving the tapestry.
func (local *Node) Kill() {
	local.blobstore.DeleteAll()
	local.server.Stop()
}

// Leave gracefully exits the Tapestry mesh.
//
// - Notify the nodes in our backpointers that we are leaving by calling NotifyLeave
// - If possible, give each backpointer a suitable alternative node from our routing table
func (local *Node) Leave() (err error) {
	// __BEGIN_TA__
	Debug.Println("Leaving tapestry")
	// xtr.NewTask("leave")
	var replacement *RemoteNode
	for i := DIGITS - 1; i >= 0; i-- {
		// Get a replacement from upper level i so that it goes into the exact same slot for nodes in backpointer
		if i != DIGITS-1 {
			routinglevel := local.table.GetLevel(i + 1)
			if len(routinglevel) > 0 {
				replacement = &(routinglevel[0])
			}
		}

		// Get the backpointers for level i
		backpointers := local.backpointers.Get(i)
		done := make(chan bool)

		// Asynchronously notify the backpointers of the leave
		for _, node := range backpointers {
			go func(node RemoteNode, replacement *RemoteNode) {
				// Notify the node of the leave
				err := node.NotifyLeaveRPC(local.node, replacement)

				// Remove the bad node so we don't select it as the replacement
				if err != nil {
					local.RemoveBadNodes([]RemoteNode{node})
				}
				// xtr.SendChannelEvent(done)
				done <- true
			}(node, replacement)
		}

		// Await completion
		for range backpointers {
			<-done
			// xtr.ReadChannelEvent(done)
		}
		Trace.Println("finished adding backpointers")
	}
	// __END_TA__
	// __BEGIN_STUDENT__
	// TODO: students should implement this
	// __END_STUDENT__
	local.blobstore.DeleteAll()
	go local.server.GracefulStop()
	return
}

// NotifyLeave occurs when another node is informing us of a graceful exit.
// - Remove references to the `from` node from our routing table and backpointers
// - If replacement is not nil or `RemoteNode{}`, add replacement to our routing table
func (local *Node) NotifyLeave(from RemoteNode, replacement *RemoteNode) (err error) {
	Debug.Printf("Received leave notification from %v with replacement node %v\n", from, replacement)

	// __BEGIN_TA__
	if local.table.Remove(from) {
		Debug.Printf("Removed %v from routing table\n", from)
	}
	if local.backpointers.Remove(from) {
		Debug.Printf("Removed %v from backpointers\n", from)
	}
	emptyRemoteNode := RemoteNode{}
	if (replacement != nil) && (*replacement != emptyRemoteNode) {
		err = local.addRoute(*replacement)
	}
	// __END_TA__
	// __BEGIN_STUDENT__
	// TODO: students should implement this
	// __END_STUDENT__
	return
}
