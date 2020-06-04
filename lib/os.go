package exifsort

import (
	"io/ioutil"
	"os"
)

func moveFile(srcPath string, dstPath string) error {
	return os.Rename(srcPath, dstPath)
}

func copyFile(src string, dst string) error {
	content, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dst, content, 0600)
	if err != nil {
		return err
	}

	return nil
}
