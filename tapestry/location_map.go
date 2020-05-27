/*
 *  Brown University, CS138, Spring 2020
 *
 *  Purpose: Defines the LocationMap type and methods for manipulating it.
 */

package tapestry

import (
	"sync"
	"time"
)

// LocationMap is struct containing objects being advertised to the tapestry.
// Object mappings are stored in the root node. An object can be advertised by multiple nodes.
// Objects time out after some amount of time if the advertising node is not heard from.
type LocationMap struct {
	data  map[string]map[RemoteNode]*time.Timer // Multimap: stores multiple nodes per key, and each node has a timeout
	mutex sync.Mutex                            // To manage concurrent access to the location map
}

// NewLocationMap creates a new objectstore.
func NewLocationMap() *LocationMap {
	m := new(LocationMap)
	m.data = make(map[string]map[RemoteNode]*time.Timer)
	return m
}

// Register registers the specified node as having advertised the key.
// Times out after the specified duration.
func (store *LocationMap) Register(key string, replica RemoteNode, timeout time.Duration) bool {
	store.mutex.Lock()

	// Get the value set for the object
	_, exists := store.data[key]
	if !exists {
		store.data[key] = make(map[RemoteNode]*time.Timer)
	}

	// Add the value to the value set
	timer, exists := store.data[key][replica]
	if !exists {
		store.data[key][replica] = store.newTimeout(key, replica, timeout)
	} else {
		timer.Reset(TIMEOUT)
	}

	store.mutex.Unlock()

	return !exists
}

// RegisterAll registers all of the provided nodes and keys.
func (store *LocationMap) RegisterAll(replicamap map[string][]RemoteNode, timeout time.Duration) {
	store.mutex.Lock()

	for key, replicas := range replicamap {
		_, exists := store.data[key]
		if !exists {
			store.data[key] = make(map[RemoteNode]*time.Timer)
		}
		for _, replica := range replicas {
			store.data[key][replica] = store.newTimeout(key, replica, timeout)
		}
	}

	store.mutex.Unlock()
}

// Unregister unregisters the specified node for the specified key.
// Returns false if the node was not registered for the key.
func (store *LocationMap) Unregister(key string, replica RemoteNode) bool {
	store.mutex.Lock()

	_, existed := store.data[key][replica]
	delete(store.data[key], replica)

	store.mutex.Unlock()

	return existed
}

// UnregisterAll unregisters all nodes that are registered for the provided key.
// Returns all replicas that were advertising the key.
func (store *LocationMap) UnregisterAll(key string) (replicas []RemoteNode) {
	store.mutex.Lock()

	replicas = slice(store.data[key])
	delete(store.data, key)

	store.mutex.Unlock()

	return
}

// Get the nodes that are advertising a given key.
func (store *LocationMap) Get(key string) (replicas []RemoteNode) {
	store.mutex.Lock()

	replicas = slice(store.data[key])

	store.mutex.Unlock()

	return
}

// GetTransferRegistrations removes and returns all objects that should be transferred to the remote node.
func (store *LocationMap) GetTransferRegistrations(local RemoteNode, remote RemoteNode) map[string][]RemoteNode {
	transfer := make(map[string][]RemoteNode)

	store.mutex.Lock()

	for key, values := range store.data {
		// Compare the first digit after the prefix
		if Hash(key).BetterChoice(remote.ID, local.ID) {
			transfer[key] = slice(values)
		}
	}

	for key := range transfer {
		delete(store.data, key)
	}

	store.mutex.Unlock()

	return transfer
}

// Utility method. Creates an expiry timer for the (key, value) pair.
func (store *LocationMap) newTimeout(key string, replica RemoteNode, timeout time.Duration) *time.Timer {
	expire := func() {
		Debug.Printf("Expiring %v for node %v\n", key, replica)

		store.mutex.Lock()

		timer, exists := store.data[key][replica]
		if exists {
			timer.Stop()
			delete(store.data[key], replica)
		}

		store.mutex.Unlock()
	}

	return time.AfterFunc(timeout, expire)
}

// Utility function to get the keys of a map
func slice(valmap map[RemoteNode]*time.Timer) (values []RemoteNode) {
	for value := range valmap {
		values = append(values, value)
	}
	return
}
