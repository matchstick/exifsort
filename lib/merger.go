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

// Merger holds the API and statistics to merge sorted directories.
type Merger struct {
	action  Action
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

// Note that when we don't specify the expression we
// assume we jut match
func mergeMatch(expression string, path string) bool {
	if expression == "" {
		return true
	}

	matched, err := regexp.MatchString(expression, path)
	if err != nil {
		fmt.Printf("%s err %s\n", path, err.Error())
		return false
	}

	return matched
}

func mergeStrToMethod(str string) Method {
	const (
		regexpYear  = `(19|[2-9][0-9])\d{2}`         // year = 1900 - 9999
		regexpMonth = `(0[1-9]|1[012])`              // month = 01 - 12
		regexpDay   = `(0[1-9]|1[0-9]|2[0-9]|3[01])` // day = 01 - 31
	)

	var (
		// We need this since windows uses "\" as a separator
		// and that is a special character for regex. It needs to be escaped.
		// Thank you QuoteMeta.
		regexSep = regexp.QuoteMeta(string(filepath.Separator))

		// Example: ^2020$
		regexpPathYear = `^` + regexpYear + `$`
		// Example: ^2020/2020_04$
		regexpPathMonth = `^` + regexpYear + regexSep + regexpYear + "_" + regexpMonth + `$`
		// Example: ^2020/2020_04/2020_04_27$
		regexpPathDay = `^` + regexpYear + regexSep + regexpYear + "_" + regexpMonth +
			regexSep + regexpYear + "_" + regexpMonth + "_" + regexpDay + `$`
	)

	if mergeMatch(regexpPathDay, str) {
		return MethodDay
	}

	if mergeMatch(regexpPathMonth, str) {
		return MethodMonth
	}

	if mergeMatch(regexpPathYear, str) {
		return MethodYear
	}

	return MethodNone
}

func mergePathValid(root string, path string) Method {
	dir := filepath.Dir(path)
	if dir == "" {
		return MethodNone
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

	return mergeStrToMethod(path)
}

// We are pretty strict on the directories we merge from and to here.
// They must fulfill several requirements:
// 1) No walk errors.
// 2) Must contain at least one media file.
// 3) Must follow the nested directory structure of:
func mergeCheck(root string) (Method, error) {
	rootMethod := MethodNone

	err := filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				errStr := fmt.Sprintf("walk err on %s with %s",
					path, err.Error())
				return errors.New(errStr)
			}

			// Don't need to scan directories
			if info.IsDir() {
				return nil
			}

			fileCategory := categorizeFile(path)
			if fileCategory == categorySkip {
				return nil
			}

			pathMethod := mergePathValid(root, path)
			if pathMethod == MethodNone {
				return fmt.Errorf("path violates method structure %s", path)
			}

			// Now we know the method of the file based on its path
			// so we compare it to other ones we found.
			switch {
			case rootMethod == MethodNone:
				// First time we found any method type for this directory
				rootMethod = pathMethod
			case rootMethod != pathMethod:
				// Cannot have more than one method in a directory
				return fmt.Errorf("%s has at least two methods %s and %s",
					root, rootMethod, pathMethod)
			case rootMethod == pathMethod:
				// We found method consistency, nothing to do.
			}

			return nil
		})

	if err != nil {
		return MethodNone, err
	}

	return rootMethod, nil
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

func (m *Merger) merge(srcPath string, srcRoot string, dstRoot string,
	action Action, logger io.Writer) error {
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

	// Finally we have everything we need to move the media
	switch action {
	case ActionCopy:
		err = copyFile(srcPath, dstPath)
	case ActionMove:
		err = moveFile(srcPath, dstPath)
	default:
		return fmt.Errorf("unknown Action %s", action)
	}

	if err != nil {
		return err
	}

	fmt.Fprintf(logger, "Merged %s to %s\n", srcPath, dstPath)
	m.storeMerged(srcPath, dstPath)

	return nil
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

			fileCategory := categorizeFile(srcFile)
			if fileCategory == categorySkip {
				return nil
			}

			if !mergeMatch(m.filter, srcFile) {
				return nil
			}

			err = m.merge(srcFile, m.srcRoot, m.dstRoot, m.action, logger)
			if err != nil {
				m.storeMergeError(srcFile, err)
				return err
			}

			return nil
		})

	return err
}

// Merge performs several actions. It checks that its src and dst directories
// are correctly sorted and determines by what method for each. If the
// directories' sort methods don't match an error is returned. Src directory is
// walked and all files are transferred to the appropriate directory in dst. New
// directories are made if necessary. If the move action was specified then
// duplicate files are removd from src. If files have the same filename and end
// up in the same dst directory they will be renamed to not collide.
func (m *Merger) Merge(logger io.Writer) error {
	srcMethod, err := mergeCheck(m.srcRoot)
	if err != nil {
		return fmt.Errorf("src dir invalid: %s", err.Error())
	}

	dstMethod, err := mergeCheck(m.dstRoot)
	if err != nil {
		return fmt.Errorf("dst dir invalid: %s", err.Error())
	}

	if srcMethod != dstMethod {
		return fmt.Errorf("src %s sorted by %s, dst %s sorted by %s",
			m.srcRoot, srcMethod, m.dstRoot, dstMethod)
	}

	return m.mergeRoots(logger)
}

// Reset will clear all previous state of a merge from the struct so it is ready
// to Merge again. src is the sorted directory to transfer from. dst is the
// sorted directory to transfer to. action is how to transfer files. filter is a
// regexp to use to filter which files to transfer
func (m *Merger) Reset(src string, dst string, action Action, filter string) {
	m.Errors = make(map[string]string)
	m.Merged = make(map[string]string)
	m.srcRoot = src
	m.dstRoot = dst
	m.action = action
	m.filter = filter
}

// NewMerger returns a Merger to execute a merge of two sorted directories.
// src is the sorted directory to transfer from. dst is the sorted directory to
// transfer to. action is how to transfer files. filter is a regexp to use to
// filter which files to transfer
func NewMerger(src string, dst string, action Action, filter string) *Merger {
	var m Merger

	m.Reset(src, dst, action, filter)

	return &m
}
