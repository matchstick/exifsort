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

func moveMedia(srcPath string, dstPath string) error {
	return os.Rename(srcPath, dstPath)
}

func copyMedia(srcPath string, dstPath string) error {
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

func ensureFullPath(path string) error {
	dirPath := filepath.Dir(path)
	return os.MkdirAll(dirPath, 0755)
}

func sortFunc(idx index, w *WalkState) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %s\n", path)
			return err
		}

		// Don't need to scan directories
		if info.IsDir() {
			return nil
		}
		// Only looking for media files that may have exif.
		if skipFileType(path) {
			w.storeSkipped()
			return nil
		}

		time, err := ExtractTime(path)
		if err != nil {
			w.storeInvalid(path, err.Error())
			w.Printf("%s\n", w.ErrStr(path, err.Error()))

			return nil
		}

		w.Printf("%s, %s\n", path, exifTimeToStr(time))
		w.storeValid(path, time)

		err = idx.Put(path, time)
		if err != nil {
			return err
		}

		return nil
	}
}

func sortTransfer(w *WalkState, m mediaMap, dst string, action int) error {
	var err error

	for newPath, oldPath := range m {
		newPath = fmt.Sprintf("%s/%s", dst, newPath)

		err = ensureFullPath(newPath)
		if err != nil {
			return err
		}

		switch action {
		case ActionCopy:
			err = copyMedia(oldPath, newPath)
		case ActionMove:
			err = moveMedia(oldPath, newPath)
		default:
			panic("Unknown action")
		}

		if err != nil {
			w.storeTransferErr(oldPath, err.Error())
			return err
		}

		w.Printf("Transferred %s\n", newPath)
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
func SortDir(src string, dst string, method int, action int, writer io.Writer) (WalkState, error) {
	w := newWalkState(writer)

	sortIndex, err := newIndex(method)
	if err != nil {
		return w, err
	}

	err = os.Mkdir(dst, 0755)
	if err != nil {
		return w, err
	}

	err = filepath.Walk(src, sortFunc(sortIndex, &w))
	if err != nil {
		return w, err
	}

	mediaMap := sortIndex.GetAll()

	err = sortTransfer(&w, mediaMap, dst, action)
	if err != nil {
		return w, err
	}

	return w, nil
}
