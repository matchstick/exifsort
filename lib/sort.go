package exifsort

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type sortError struct {
	prob string
}

func (e sortError) Error() string {
	return e.prob
}

func (s *Sorter) moveMedia(srcPath string, dstPath string) error {
	return os.Rename(srcPath, dstPath)
}

func (s *Sorter) copyMedia(srcPath string, dstPath string) error {
	srcStat, err := os.Stat(srcPath)
	if err != nil {
		return err
	}

	if !srcStat.Mode().IsRegular() {
		errStr := fmt.Sprintf("%s is not a regular file", srcPath)
		return sortError{errStr}
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)

	return err
}

func (s *Sorter) ensureFullPath(path string) error {
	dirPath := filepath.Dir(path)
	return os.MkdirAll(dirPath, 0755)
}

type Sorter struct {
	idx       index
	idxErrors map[string]error
	xfrErrors map[string]error
}

// We don't check if you have a path duplicate.
func (s *Sorter) storeIndexError(path string, err error) {
	s.idxErrors[path] = err
}

// We don't check if you have a path duplicate.
func (s *Sorter) storeTransferError(path string, err error) {
	s.xfrErrors[path] = err
}

func (s *Sorter) IndexErrors() map[string]error {
	return s.idxErrors
}

func (s *Sorter) TransferErrors() map[string]error {
	return s.xfrErrors
}

func NewSorter(w WalkState, method int) (*Sorter, error) {
	var s Sorter
	s.xfrErrors = make(map[string]error)
	s.idxErrors = make(map[string]error)

	idx, err := newIndex(method)
	if err != nil {
		return nil, err
	}

	s.idx = idx

	for path, time := range w.Data() {
		err = s.idx.Put(path, time)
		if err != nil {
			s.storeIndexError(path, err)
		}
	}

	return &s, nil
}

func (s *Sorter) Transfer(dst string, action int, logger io.Writer) error {
	err := os.Mkdir(dst, 0755)
	if err != nil {
		return err
	}

	mediaMap := s.idx.GetAll()

	for newPath, oldPath := range mediaMap {
		newPath = fmt.Sprintf("%s/%s", dst, newPath)

		err = s.ensureFullPath(newPath)
		if err != nil {
			return err
		}

		switch action {
		case ActionCopy:
			err = s.copyMedia(oldPath, newPath)
		case ActionMove:
			err = s.moveMedia(oldPath, newPath)
		default:
			panic("Unknown action")
		}

		if err != nil {
			s.storeTransferError(oldPath, err)
			return err
		}

		fmt.Fprintf(logger, "Transferred %s\n", newPath)
	}

	return nil
}

// SortDir examines the contents of file with acceptable suffixes in the src
// directory and transfer the file to the dst directory. The structure of
// the dst directory is specified by 'method'. The action of transfer is
// specified by 'action'. It returns WalkState gathered as a return value.
//
// SortDir only scans media files listed as constants as documented, other
// files are skipped.
//
// writer is where to write output while scanning. nil for none.
