package exifsort

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/udhos/equalfile"
)

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

// Routine to test if the argument filename will collide with any
// other filenames for the caller context.
type doesCollideFunc func(filename string) bool

// When you find a collision you add a counter to the filename.
// So <name>.jpg => <name>_#.jpg the number increments as it may have
// multiple collisions. This way we can create a new unique name.
// We accept a doesCollideFunc to have the context for collisions be clean.
func collisionName(base string, doesCollide doesCollideFunc) string {
	var newName string

	extension := filepath.Ext(base)
	prefix := strings.TrimRight(base, extension)

	// Now we keep trying until we create a name that won't collide
	for counter := 0; true; counter++ {
		newName = fmt.Sprintf("%s_%d%s", prefix, counter, extension)

		if !doesCollide(newName) {
			break
		}
	}

	return newName
}
