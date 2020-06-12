package exifsort

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/udhos/equalfile"
)

// We need to indicate when we have found a file that is an exact duplicate
// (filename and contents) of another.
type duplicateError struct {
	src string
	dst string
	Err error
}

func (e *duplicateError) Error() string {
	return fmt.Sprintf("%s is a duplicate of %s", e.src, e.dst)
}

func (e *duplicateError) Unwrap() error { return e.Err }

func moveFile(srcPath string, dstPath string) error {
	return os.Rename(srcPath, dstPath)
}

func copyFile(src string, dst string) error {
	content, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dst, content, 0600)
	if err != nil {
		return err
	}

	return nil
}

func isEqual(lhs string, rhs string) (bool, error) {
	// Check for same contents
	cmp := equalfile.New(nil, equalfile.Options{})

	equal, err := cmp.CompareFile(lhs, rhs)
	if err != nil {
		return false, err
	}

	return equal, nil
}

// Routine returns what path that collides with the argument
type collisionNameFunc func(filename string) string

// uniqueName's purpose is to return to the caller a filename that is unique
// among the context provided in the argument func. If it discovers a file that
// has the same filename and the same contents it will return an error. We
// don't want to add a file that is a duplicate.
//
// The routine achieves ths goal by when finding a collision it reconstructs
// the filename with a counter as part of the name.
//
// So <name>.jpg => <name>_#.jpg. The number increments as it may have
// multiple collisions. This way we can create a new unique name.
// We accept a function to determine if the filenames collide with the caller's
// file set.
func uniqueName(path string, doesCollide collisionNameFunc) (string, error) {
	var filename = filepath.Base(path)

	extension := filepath.Ext(filename)
	prefix := strings.TrimRight(filename, extension)

	for counter := 0; true; counter++ {
		// Test for unique filename
		collisionPath := doesCollide(filename)
		if collisionPath == "" {
			// There is no collisionPath so filename is unique
			break
		}

		sameContents, err := isEqual(path, collisionPath)
		if err != nil {
			// Some error in comparison
			return "", err
		}

		if sameContents {
			// If filename and contents are the same then
			// no need to add this file, it is not unique.
			var err duplicateError
			err.src = path
			err.dst = collisionPath

			return "", &err
		}

		// Try a new filename then
		filename = fmt.Sprintf("%s_%d%s", prefix, counter, extension)
	}

	return filename, nil
}
