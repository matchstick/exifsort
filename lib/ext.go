package exifsort

import (
	"path/filepath"
	"strings"
)

// Files extensions that are processed and not skipped.
const (
	ExtensionBMP = iota
	ExtensionCR2
	ExtensionDNG
	ExtensionGIF
	ExtensionJPEG
	ExtensionJPG
	ExtensionNEF
	ExtensionPNG
	ExtensionPSD
	ExtensionRAF
	ExtensionRAW
	ExtensionTIF
	ExtensionTIFF
)

func extensionMap() map[string]int {
	// We are going to do this check a lot so let's use a map.
	return map[string]int{
		".bmp":  ExtensionBMP,
		".cr2":  ExtensionCR2,
		".dng":  ExtensionDNG,
		".gif":  ExtensionGIF,
		".jpeg": ExtensionJPEG,
		".jpg":  ExtensionJPG,
		".nef":  ExtensionNEF,
		".png":  ExtensionPNG,
		".psd":  ExtensionPSD,
		".raf":  ExtensionRAF,
		".raw":  ExtensionRAW,
		".tif":  ExtensionTIF,
		".tiff": ExtensionTIFF,
	}
}

func skipFileType(path string) (string, bool) {
	// All comparisons are lower case as case don't matter
	path = strings.ToLower(path)

	// Running this on a synology results in the file server
	// creating all these useless media files. We want to skip
	// them.
	switch {
	case strings.Contains(path, "@eadir"):
		return "", true
	case strings.Contains(path, "@syno"):
		return "", true
	case strings.Contains(path, "synofile_thumb"):
		return "", true
	}

	extension := filepath.Ext(path)
	// no extension to check so we skip
	if extension == "" {
		return "", true
	}

	_, inExtMap := extensionMap()[extension]

	return extension, !inExtMap
}
