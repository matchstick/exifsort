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

const (
	yearRe  = `(19|[2-9][0-9])\d{2}`         // year = 1900 - 9999
	monthRe = `(0[1-9]|1[012])`              // month = 01 - 12
	dayRe   = `(0[1-9]|1[0-9]|2[0-9]|3[01])` // day = 01 - 31
)

func mergePathValid(root string, path string, method int) bool {
	dir := filepath.Dir(path)
	if dir == "" {
		return false
	}

	// We want the replace below to strip out the "/" for us
	// So we add it here
	if strings.LastIndexByte(root, filepath.Separator) != len(root)-1 {
		// We do this as the separator is a rune for windows.
		runeRoot := []rune(root)
		runeRoot = append(runeRoot, filepath.Separator)
		root = string(runeRoot)
	}

	// get the time based paths.
	path = strings.Replace(dir, root, "", 1)

	var matchStr string

	var regexSep = regexp.QuoteMeta(string(filepath.Separator))

	switch method {
	case MethodYear:
		matchStr = yearRe
	case MethodMonth:
		matchStr = yearRe + regexSep +
			yearRe + "_" + monthRe
	case MethodDay:
		matchStr = yearRe + regexSep +
			yearRe + "_" + monthRe + regexSep +
			yearRe + "_" + monthRe + "_" + dayRe
	default:
		return false
	}

	matchStr = `^` + matchStr + `$`

	fmt.Printf("MatchStr: %s\n\n", matchStr)

	isMatch, err := regexp.MatchString(matchStr, path)
	if err != nil {
		return false
	}

	return isMatch
}

// We are pretty strict on the directories we merge from and to here.
// They must fulfill several requirements:
// 1) No walk errors.
// 2) Must contain at least one media file.
// 3) Must follow the nested directory structure of:
func MergeCheck(root string, method int) error {
	err := filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				errStr := fmt.Sprintf("Walk Err on %s with %s",
					path, err.Error())
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

func merge(srcFile string, srcRoot string, dstRoot string, action int) error {
	dstFile := strings.Replace(srcFile, srcRoot, dstRoot, 1)

	switch action {
	case ActionCopy:
		return copyFile(srcFile, dstFile)
	case ActionMove:
		return moveFile(srcFile, dstFile)
	default:
		errStr := fmt.Sprintf("Unknown Action %d", action)
		return &mergeErr{errStr}
	}
}

func Merge(srcRoot string, dstRoot string, method int, logger io.Writer) error {
	err := filepath.Walk(srcRoot,
		func(srcFile string, info os.FileInfo, err error) error {
			if err != nil {
				errStr := fmt.Sprintf("Err on %s with %s",
					srcFile, err.Error())
				return &mergeErr{errStr}
			}

			// Don't need to scan directories
			if info.IsDir() {
				return nil
			}

			_, skip := skipFileType(srcFile)
			if skip {
				return nil
			}

			err = merge(srcFile, srcRoot, dstRoot, method)
			if err != nil {
				return err
			}

			return nil
		})

	return err
}
