package exifsort

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type mergeErr struct {
	prob string
}

func (m mergeErr) Error() string {
	return m.prob
}

func mergePathValid(root string, path string, method int) bool {
	dir := filepath.Dir(path)
	if dir == "" {
		return false
	}

	// get the time based paths.
	pieces := strings.TrimLeft(root, dir)

	isMatch := false

	var err error

	switch method {
	case MethodYear:
		isMatch, err = regexp.MatchString(`\d\d\d\d\/`, pieces)
	case MethodMonth:
		isMatch, err = regexp.MatchString(`\d\d\d\d\/\d\d\d\d_\d\d`, pieces)
	case MethodDay:
		isMatch, err = regexp.MatchString(`\d\d\d\d\/\d\d\d\d_\d\d\/\d\d\d\d_\d\d_\d\d`, pieces)
	default:
		return false
	}

	if err != nil {
		return false
	}

	return isMatch
}

// We are pretty strict on the directoires we merge from and to here.
// They must fulfill several requirements:
// 1) No walk errors.
// 2) Must contain at least one media file.
// 3) Must follow the nested directory structure of:
// TODO.
func mergeCheck(root string, method int, logger io.Writer) error {
	err := filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				errStr := fmt.Sprintf("Walk Err on %s with %s", path, err.Error())
				return &mergeErr{errStr}
			}

			// Don't need to scan directories
			if info.IsDir() {
				return nil
			}

			_, skip := skipFileType(path)
			if skip {
				return nil
			}

			if !mergePathValid(root, path, method) {
				errStr := fmt.Sprintf("Illegal Path %s", path)
				return &mergeErr{errStr}
			}

			return nil
		})

	return err
}
