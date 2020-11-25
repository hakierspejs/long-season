// static package implements methods for manipulating files
// embedded into long-season binary.
package static

import "os"

// Open returns content of file embedded into
// long-season binary. Returns error if given file
// does not exists.
func Open(path string) ([]byte, error) {
	b, ok := files[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return b, nil
}
