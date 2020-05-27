/*
 *  Brown University, CS138, Spring 2020
 *
 *  Purpose: sets up several loggers and provides utility methods for printing
 *  tapestry structures.
 */

package tapestry

//__BEGIN_TA__
import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	// xtr "github.com/brown-csci1380/tracing-framework-go/xtrace/client"
	"google.golang.org/grpc/grpclog"
)

//__END_TA__
/*__BEGIN_STUDENT__
import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"google.golang.org/grpc/grpclog"
	// Uncomment for xtrace
	// xtr "github.com/brown-csci1380/tracing-framework-go/xtrace/client"
)
__END_STUDENT__*/

// Debug is optional debug logger
var Debug *log.Logger

// Out logs to stdout
var Out *log.Logger

// Error logs to stderr
var Error *log.Logger

// Trace is used for xtrace
var Trace *log.Logger

// Initialize the loggers
func init() {
	Debug = log.New(ioutil.Discard, "", log.Ltime|log.Lshortfile)
	Trace = log.New(ioutil.Discard, "", log.Lshortfile)
	//__BEGIN_TA__
	//__END_TA__
	/*__BEGIN_STUDENT__
	// Uncomment for xtrace
	// Trace.SetOutput(xtr.MakeWriter())
	__END_STUDENT__ */
	Out = log.New(os.Stdout, "", log.Ltime|log.Lshortfile)
	Error = log.New(os.Stderr, "ERROR: ", log.Ltime|log.Lshortfile)
	grpclog.SetLogger(log.New(ioutil.Discard, "", log.Ltime))
}

// SetDebug turns debug on or off
func SetDebug(enabled bool) {
	if enabled {
		Debug.SetOutput(os.Stdout)
		// uncomment for xtrace
		// Debug.SetOutput(xtr.MakeWriter(os.Stdout))
	} else {
		Debug.SetOutput(ioutil.Discard)
	}
}

// RoutingTableToString stringifies the routing table
func (tapestry *Node) RoutingTableToString() string {
	var buffer bytes.Buffer
	table := tapestry.table
	fmt.Fprintf(&buffer, "RoutingTable for node %v\n", tapestry.node)
	id := table.local.ID.String()
	for i, row := range table.rows {
		for j, slot := range row {
			for _, node := range slot {
				fmt.Fprintf(&buffer, " %v%v  %v: %v %v\n", id[:i], strings.Repeat(" ", DIGITS-i+1), Digit(j), node.Address, node.ID.String())
			}
		}
	}

	return buffer.String()
}

// PrintRoutingTable prints the routing table
func (tapestry *Node) PrintRoutingTable() {
	fmt.Println(tapestry.RoutingTableToString())
}

// LocationMapToString stringifies the location map
func (tapestry *Node) LocationMapToString() string {
	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "LocationMap for node %v\n", tapestry.node)
	for key, values := range tapestry.locationsByKey.data {
		fmt.Fprintf(&buffer, " %v: %v\n", key, slice(values))
	}

	return buffer.String()
}

// PrintLocationMap prints the location map
func (tapestry *Node) PrintLocationMap() {
	fmt.Printf(tapestry.LocationMapToString())
}

// BackpointersToString stringifies the backpointers
func (tapestry *Node) BackpointersToString() string {
	var buffer bytes.Buffer
	bp := tapestry.backpointers
	fmt.Fprintf(&buffer, "Backpointers for node %v\n", tapestry.node)
	for i, set := range bp.sets {
		for _, node := range set.Nodes() {
			fmt.Fprintf(&buffer, " %v %v: %v\n", i, node.Address, node.ID.String())
		}
	}

	return buffer.String()
}

// PrintBackpointers prints the backpointers
func (tapestry *Node) PrintBackpointers() {
	fmt.Printf(tapestry.BackpointersToString())
}

// BlobStoreToString stringifies the blob store
func (tapestry *Node) BlobStoreToString() string {
	var buffer bytes.Buffer
	for k := range tapestry.blobstore.blobs {
		fmt.Fprintln(&buffer, k)
	}
	return buffer.String()
}

// PrintBlobStore prints the blobstore
func (tapestry *Node) PrintBlobStore() {
	fmt.Printf(tapestry.BlobStoreToString())
}
