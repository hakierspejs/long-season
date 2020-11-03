// Package update contains utilities for updating various
// data structures.
package update

// String returns updated string if it's length
// is greater than zero, otherwise returns old.
func String(old, updated string) string {
	if len(updated) > 0 {
		return updated
	}
	return old
}

// Bytes returns updated bytes slice if it's length
// is greater than zero, otherwise returns old.
func Bytes(old, updated []byte) []byte {
	if len(updated) > 0 {
		return updated
	}
	return old
}

// NullableBool returns updated value if it is not
// null, otherwise returns old.
func NullableBool(old bool, updated *bool) bool {
	if updated != nil {
		return *updated
	}
	return old
}
