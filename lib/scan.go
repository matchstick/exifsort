package exifsort

import (
	"os"
	"path/filepath"
)

func scanFunc(w *WalkState) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			w.storeInvalid(path, err.Error())
			w.walkPrintf("%s\n", w.ErrStr(path, err.Error()))

			return nil
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

		return nil
	}
}

// ScanDir will examine the contents of every file in the src directory and
// print it's time of creation as stored by exifdata as it scans. It returns
// WalkState gathered as a return value.
//
// ScanDir only scans media files listed as constants as documented, other
// files are skipped.
//
// If doPrint is set to false it will not print while scanning.
func ScanDir(src string, doPrint bool) WalkState {
	w := newWalkState(doPrint)

	// scanFunc never returns an error
	// We don't want to walk for an hour and then fail on one error.
	// Consult the walkstate for errors.
	_ = filepath.Walk(src, scanFunc(&w))

	return w
}
