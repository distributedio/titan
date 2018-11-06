package db

// Lease is an object that can be associated with other objects
// those can share the same ttl of the lease
type Lease struct {
	Object
	TouchedAt int64
}
