package exifsort

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type sortError struct {
	prob string
}

func (e sortError) Error() string {
	return e.prob
}

// Sorter is your API to perform sorting actions after a scan.
//
// Sorting is composed of two stages. The first is "indexing", where the scanned
// data is organized by the specified method. The second is "transferring", where
// the media is transferred from the src to the dst directory.
//
// It holds the index of sorted media and errors found in constructing or
// transferring it.
type Sorter struct {
	idx            index
	IndexErrors    map[string]string
	TransferErrors map[string]string
	Duplicates     []string
}

func (s *Sorter) ensureFullPath(path string) error {
	dirPath := filepath.Dir(path)
	return os.MkdirAll(dirPath, 0755)
}

func (s *Sorter) storeDuplicate(path string) {
	s.Duplicates = append(s.Duplicates, path)
}

// We don't check if you have a path duplicate.
func (s *Sorter) storeIndexError(path string, err error) {
	s.IndexErrors[path] = err.Error()
}

// We don't check if you have a path duplicate.
func (s *Sorter) storeTransferError(path string, err error) {
	s.TransferErrors[path] = err.Error()
}

// Performs the transfer after indexing.
// Transfer will fail if dst directory does not exist and is not accessible.
func (s *Sorter) Transfer(dst string, action int, logger io.Writer) error {
	if action != ActionCopy && action != ActionMove {
		errStr := fmt.Sprintf("Invalid action %d\n", action)
		return &sortError{errStr}
	}

	info, err := os.Stat(dst)
	if err != nil || !info.IsDir() {
		errStr := fmt.Sprintf("Error: No Output dir: %s", dst)
		return &sortError{errStr}
	}

	// Let's get rid of all the duplciates we know of before we transfer.
	for _, toRemove := range s.Duplicates {
		err := os.Remove(toRemove)
		if err != nil {
			s.storeTransferError(toRemove, err)
		}
	}

	mediaMap := s.idx.GetAll()

	for newPath, oldPath := range mediaMap {
		newPath = filepath.Join(dst, newPath)

		err = s.ensureFullPath(newPath)
		if err != nil {
			return err
		}

		switch action {
		case ActionCopy:
			err = copyFile(oldPath, newPath)
		case ActionMove:
			err = moveFile(oldPath, newPath)
		}

		if err != nil {
			s.storeTransferError(oldPath, err)
			return err
		}

		fmt.Fprintf(logger, "Transferred %s\n", newPath)
	}

	return nil
}

func (s *Sorter) Reset(scanner Scanner, method int) error {
	s.IndexErrors = make(map[string]string)
	s.TransferErrors = make(map[string]string)

	idx, err := newIndex(method)
	if err != nil {
		return err
	}

	s.idx = idx

	for path, time := range scanner.Data {
		err = s.idx.Put(path, time)
		if err == nil {
			continue
		}

		// Do we have a duplicate?
		if strings.Contains(err.Error(), "duplicate") {
			s.storeDuplicate(path)
			continue
		}

		s.storeIndexError(path, err)
	}

	return nil
}

// Creates the sorter based on the WalkState generated by a scan and the method
// desired to sort.
//
// The structure of how it will organize the the dst directory is specified by
// 'method'. This routine will index via the method speficied.
func NewSorter(scanner Scanner, method int) (*Sorter, error) {
	var s Sorter

	err := s.Reset(scanner, method)
	if err != nil {
		return nil, err
	}

	return &s, nil
}
