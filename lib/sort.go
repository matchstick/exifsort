package exifsort

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func moveMedia(srcPath string, dstPath string) error {
	return os.Rename(srcPath, dstPath)
}

func copyMedia(srcPath string, dstPath string) error {
	srcStat, err := os.Stat(srcPath)
	if err != nil {
		return err
	}

	if !srcStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", srcPath)
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
			w.walkPrintf("%s\n", w.ErrStr(path, err.Error()))

			return nil
		}

		w.walkPrintf("%s, %s\n", path, exifTimeToStr(time))
		w.storeValid()

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

		w.walkPrintf("Transferred %s\n", newPath)
	}

	return nil
}

// SortDir examines the contents of file with acceptable suffixes in the src
// directory and transfer the file to the dst directory. The structure of
// the dst directory is specified by 'method'. The action of transfer is
// specified by 'action'.
//
// SortDir only scans media files listed in the mediaSuffixMap, other files are
// skipped.
//
// When 'summarize' is set to 'true' it will print a summary of stats when
// completed.
//
// If doPrint is set to false it will not print while scanning.
func SortDir(src string, dst string, method int, action int, doPrint bool) (WalkState, error) {
	w := newWalkState(doPrint)
	sortIndex := newIndex(method)

	err := os.Mkdir(dst, 0755)
	if err != nil {
		return w, fmt.Errorf("cannot make directory %s", dst)
	}

	err = filepath.Walk(src, sortFunc(sortIndex, &w))
	if err != nil {
		return w, fmt.Errorf("sort Walk Error (%s)", err.Error())
	}

	mediaMap := sortIndex.GetAll()

	err = sortTransfer(&w, mediaMap, dst, action)
	if err != nil {
		return w, err
	}

	return w, nil
}
