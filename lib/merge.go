package exifsort

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

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

	// We need this since windows uses "\" as a separator
	// and that is a special character for regex. It needs to be escaped.
	// Thank you QuoteMeta.
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
				return errors.New(errStr)
			}

			// Don't need to scan directories
			if info.IsDir() {
				return nil
			}

			if !mergePathValid(root, path, method) {
				errStr := fmt.Sprintf("Illegal Path %s", path)
				return errors.New(errStr)
			}

			return nil
		})

	return err
}

func merge(srcFile string, srcRoot string, dstRoot string, action int) error {
	// Remove the root
	filePath := strings.Replace(srcFile, srcRoot, "", 1)

	// The directories we are going to put the file into
	dstDir := filepath.Join(dstRoot, filepath.Dir(filePath))

	dirEntries, err := ioutil.ReadDir(dstDir)
	if err != nil {
		return errors.New(err.Error())
	}

	entryMap := make(map[string]string)

	for _, entry := range dirEntries {
		baseName := filepath.Base(entry.Name())
		fullPath := filepath.Join(dstDir, entry.Name())
		entryMap[baseName] = fullPath
	}

	dstFile, err := uniqueName(srcFile, func(filename string) string {
		return entryMap[filename]
	})

	// If we have a duplicate error that the file is a duplicate.
	// If we are moving files then remove the source.
	var dupErr *duplicateError
	if errors.As(err, &dupErr) && action == ActionMove {
		return os.Remove(dupErr.src)
	}

	// If we have a real error we need to stop.
	if err != nil {
		return err
	}

	dstFile = filepath.Join(dstDir, dstFile)

	switch action {
	case ActionCopy:
		return copyFile(srcFile, dstFile)
	case ActionMove:
		return moveFile(srcFile, dstFile)
	default:
		errStr := fmt.Sprintf("Unknown Action %d", action)
		return errors.New(errStr)
	}
}

func Merge(srcRoot string, dstRoot string, action int, logger io.Writer) error {
	err := filepath.Walk(srcRoot,
		func(srcFile string, info os.FileInfo, err error) error {
			if err != nil {
				errStr := fmt.Sprintf("Err on %s with %s",
					srcFile, err.Error())
				return errors.New(errStr)
			}

			// Don't need to scan directories
			if info.IsDir() {
				return nil
			}

			if skipFileType(srcFile) {
				return nil
			}

			err = merge(srcFile, srcRoot, dstRoot, action)
			if err != nil {
				return err
			}

			return nil
		})

	return err
}
