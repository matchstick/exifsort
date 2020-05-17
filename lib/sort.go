package exifSort

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var sortIndex index

func moveMedia(srcPath string, dstPath string) error {
	return os.Rename(srcPath, dstPath)
}

func copyMedia(srcPath string, dstPath string) (int64, error) {
	srcStat, err := os.Stat(srcPath)
	if err != nil {
		return 0, err
	}

	if !srcStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", srcPath)
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return 0, err
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return 0, err
	}
	defer dst.Close()
	nBytes, err := io.Copy(dst, src)
	return nBytes, err
}

func ensureFullPath(path string) error {
	dirPath := filepath.Dir(path)
	return os.MkdirAll(dirPath, 0755)
}

func sortSummary(summarize bool) {
	if !summarize {
		return
	}
	fmt.Printf("Sort Summary\n")
}

func sortFunc(path string, info os.FileInfo, err error) error {
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
		walkState.storeSkipped()
		return nil
	}

	time, err := ExtractTime(path)
	if err != nil {
		walkState.storeInvalid(path, err.Error())
		walkState.walkPrintf("%s\n", walkErrMsg(path, err.Error()))
		return nil
	}

	walkState.storeValid()
	err = sortIndex.Put(path, time)
	if err != nil {
		return err
	}
	return nil
}

func sortTransfer(m mediaMap, dst string, action int) error {
	for newPath, oldPath := range m {
		newPath = fmt.Sprintf("%s/%s", dst, newPath)
		err := ensureFullPath(newPath)
		if err != nil {
			return err
		}
		switch action {
		case ACTION_COPY:
			_, err := copyMedia(oldPath, newPath)
			if err != nil {
				return err
			}
		case ACTION_MOVE:
			err := moveMedia(oldPath, newPath)
			if err != nil {
				return err
			}
		default:
			panic("Unknown action")
		}
	}
	return nil
}

// SortDir examines the contents of file with acceptable suffixes in the src
// directory and transfer the file to the dst directory. The structure of
// the dst directory is specifed by 'method'. The action of transfer is
// specified by 'action'.
//
// SortDir only scans media files listed in the mediaSuffixMap, other files are
// skipped.
//
// When 'summarize' is set to 'true' it will print a summary of stats when
// completed.
//
// If doPrint is set to false it will not print while scanning.
func SortDir(src string, dst string, method int, action int, summarize bool, doPrint bool) error {
	walkState.init(doPrint)
	sortIndex = createIndex(method)

	err := os.Mkdir(dst, 0755)
	if err != nil {
		return fmt.Errorf("Cannot make directory %s\n", dst)
	}

	err = filepath.Walk(src, sortFunc)
	if err != nil {
		return fmt.Errorf("Sort Walk Error (%s)\n", err.Error())
	}

	mediaMap := sortIndex.GetAll()
	err = sortTransfer(mediaMap, dst, action)
	if err != nil {
		return err
	}
	sortSummary(summarize)
	return nil
}
