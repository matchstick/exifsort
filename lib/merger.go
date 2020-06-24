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

type Merger struct {
	action  Action
	method  Method
	srcRoot string
	dstRoot string
	filter  string
	Merged  map[string]string
	Errors  map[string]string
	Removed []string
}

func (m *Merger) storeMergeRemoved(path string) {
	m.Removed = append(m.Removed, path)
}

func (m *Merger) storeMergeError(path string, err error) {
	m.Errors[path] = err.Error()
}

func (m *Merger) storeMerged(src string, dst string) {
	m.Merged[dst] = src
}

const (
	regexpYear  = `(19|[2-9][0-9])\d{2}`         // year = 1900 - 9999
	regexpMonth = `(0[1-9]|1[012])`              // month = 01 - 12
	regexpDay   = `(0[1-9]|1[0-9]|2[0-9]|3[01])` // day = 01 - 31
)

/*
var (
		// We need this since windows uses "\" as a separator
	// and that is a special character for regex. It needs to be escaped.
	// Thank you QuoteMeta.
	regexSep = regexp.QuoteMeta(string(filepath.Separator))

regexpPathYear  = regexpYear
	regexpPathMonth = regexpPathYear + regexSep + regexpYear + "_" + regexpMonth
	regexpPathDay   = regexpPathYear + regexpPathMonth + regexSep +
		regexpYear + "_" + regexpMonth + "_" + regexpDay
)

func regexMatch(expression string, str string) bool {
	str = `^` + str + `$`
	isMatch, err := regexp.MatchString(expression, str)
	if err != nil {
		return false
	}

	return isMatch
}

func strToMethod(str string) Method {
	switch {
	case regexMatch(regexpPathYear, str):
		return MethodYear
	case regexMatch(regexpPathMonth, str):
		return MethodYear
	case regexMatch(regexpPathDay, str):
		return MethodYear
	default:
		return MethodNone
	}
}
*/

func mergePathValid(root string, path string, method Method) bool {
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

	// We need this since windows uses "\" as a separator
	// and that is a special character for regex. It needs to be escaped.
	// Thank you QuoteMeta.
	regexSep := regexp.QuoteMeta(string(filepath.Separator))

	var matchStr string

	switch method {
	case MethodYear:
		matchStr = regexpYear
	case MethodMonth:
		matchStr = regexpYear + regexSep +
			regexpYear + "_" + regexpMonth
	case MethodDay:
		matchStr = regexpYear + regexSep +
			regexpYear + "_" + regexpMonth + regexSep +
			regexpYear + "_" + regexpMonth + "_" + regexpDay
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
func mergeCheck(root string, method Method) error {
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

func isMatch(matchStr string, path string) bool {
	if matchStr == "" {
		return true
	}

	matched, err := regexp.MatchString(matchStr, path)
	if err != nil {
		fmt.Printf("%s err %s\n", path, err.Error())
		return false
	}

	return matched
}

func (m *Merger) mergeDuplicate(err error, action Action) error {
	// Is this error a duplicate file?
	// If not a duplicate error just propagate it
	var dupErr *duplicateError
	if !errors.As(err, &dupErr) {
		m.storeMergeError(dupErr.src, dupErr)
		return err
	}

	// If we are not moving files then we should just do nothing
	if action != ActionMove {
		return nil
	}

	// We have a duplicate so we have to remove it
	// Hopefully there is no os problem with doing so.
	m.storeMergeRemoved(dupErr.src)

	return os.Remove(dupErr.src)
}

func (m *Merger) merge(srcPath string, srcRoot string, dstRoot string, action Action) error {
	// Remove the root but this is not the basename just what is between
	// root and base
	filePath := strings.Replace(srcPath, srcRoot, "", 1)

	// The directory we are going to put the file into
	dstDir := filepath.Join(dstRoot, filepath.Dir(filePath))

	dirEntries, err := ioutil.ReadDir(dstDir)

	var dstPath string

	switch {
	case os.IsNotExist(err):
		// mkdir as we need to
		err = os.MkdirAll(dstDir, 0777)
		if err != nil {
			return err
		}

		// We know dstPath is unique, first file in the directory we just made
		srcBase := filepath.Base(srcPath)
		dstPath = filepath.Join(dstDir, srcBase)
	case err != nil:
		// We have an error
		return errors.New(err.Error())
	default:
		// Here we know the directory pre-exists we have the entries.
		// So we need to find a unique filename and build a path for
		// the target.
		entryMap := make(map[string]string)

		// Collect the existing names
		for _, entry := range dirEntries {
			baseName := filepath.Base(entry.Name())
			fullPath := filepath.Join(dstDir, entry.Name())
			entryMap[baseName] = fullPath
		}

		// Find a new one based on ours
		dstBase, err := uniqueName(srcPath, func(filename string) string {
			return entryMap[filename]
		})

		// We got an error it's either a duplicate or a real problem.
		if err != nil {
			return m.mergeDuplicate(err, action)
		}

		dstPath = filepath.Join(dstDir, dstBase)
	}

	m.storeMerged(srcPath, dstPath)

	// Finally we have everything we need to move the media
	switch action {
	case ActionCopy:
		return copyFile(srcPath, dstPath)
	case ActionMove:
		return moveFile(srcPath, dstPath)
	default:
		return fmt.Errorf("unknown Action %s", action)
	}
}

func (m *Merger) mergeRoots(logger io.Writer) error {
	err := filepath.Walk(m.srcRoot,
		func(srcFile string, info os.FileInfo, err error) error {
			if err != nil {
				m.storeMergeError(srcFile, err)
				return fmt.Errorf("walk on %s with %s", srcFile, err.Error())
			}

			// Don't need to scan directories
			if info.IsDir() {
				return nil
			}

			if !isMatch(m.filter, srcFile) {
				return nil
			}

			err = m.merge(srcFile, m.srcRoot, m.dstRoot, m.action)
			if err != nil {
				m.storeMergeError(srcFile, err)
				return err
			}

			return nil
		})

	return err
}

func (m *Merger) Merge(logger io.Writer) error {
	err := mergeCheck(m.srcRoot, m.method)
	if err != nil {
		return fmt.Errorf("src dir invalid: %s", err.Error())
	}

	err = mergeCheck(m.dstRoot, m.method)
	if err != nil {
		return fmt.Errorf("dst dir invalid: %s", err.Error())
	}

	return m.mergeRoots(logger)
}

func (m *Merger) Reset(srcRoot string, dstRoot string, action Action, method Method, filter string) {
	m.Errors = make(map[string]string)
	m.Merged = make(map[string]string)
	m.srcRoot = srcRoot
	m.dstRoot = dstRoot
	m.action = action
	m.method = method
	m.filter = filter
}

func NewMerger(srcRoot string, dstRoot string, action Action, method Method,
	filter string) *Merger {
	var m Merger

	m.Reset(srcRoot, dstRoot, action, method, filter)

	return &m
}
