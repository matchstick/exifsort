package exifsort

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestMergeCheckGood(t *testing.T) {
	for _, goodMethod := range Methods() {
		goodMethod := goodMethod

		t.Run(goodMethod.String(), func(t *testing.T) {
			t.Parallel()

			td := newTestDir(t, goodMethod, fileNoDefault)

			src := td.buildRoot()
			defer os.RemoveAll(src)

			dst := td.buildSortedDir(src, "dst", ActionCopy)

			method, err := mergeCheck(dst)
			if err != nil {
				t.Errorf("Err %s, method %s\n", err.Error(), method)
			}

			if method != goodMethod {
				t.Errorf("Err %s, method %s should be %s\n", err.Error(), method,
					goodMethod)
			}

			os.RemoveAll(dst)
		})
	}
}

func TestMergeCheckBad(t *testing.T) {
	for _, goodMethod := range Methods() {
		goodMethod := goodMethod

		t.Run(goodMethod.String(), func(t *testing.T) {
			t.Parallel()

			td := newTestDir(t, goodMethod, fileNoDefault)

			src := td.buildRoot()
			defer os.RemoveAll(src)

			dst := td.buildSortedDir(src, "dst", ActionCopy)

			badFilePath := filepath.Join(dst, "testfile.CR2")
			message := []byte("Hello, Gophers!")

			_ = ioutil.WriteFile(badFilePath, message, 0600)

			method, err := mergeCheck(dst)
			if err == nil {
				t.Errorf("Unexpected Success method %s\n", method)
			}

			if method != MethodNone {
				t.Errorf("Method %s is not MethodNone as expected\n", method)
			}

			os.RemoveAll(badFilePath)
			os.RemoveAll(dst)
		})
	}
}

func testSrcTotal(tdSrc *testdir, tdDst *testdir, action Action) int {
	var srcTotal int

	switch action {
	case ActionCopy:
		// src dir should have all its media untouched
		srcTotal = tdSrc.numTotal()
	case ActionMove:
		// src dir should have all media merged and be empty
		srcTotal = 0
	}

	return srcTotal
}

func testDstTotal(tdSrc *testdir, tdDst *testdir, dup bool) int {
	if dup {
		// Destination should have data only from dst as src is all
		// duplicates
		return tdDst.numData
	}

	// Destination should have all data from both sources
	return tdDst.numData + tdSrc.numData
}

func testMergeResults(t *testing.T, tdSrc *testdir, tdDst *testdir,
	fromDir string, toDir string, action Action, dup bool) error {
	srcTotal := testSrcTotal(tdSrc, tdDst, action)

	err := countFiles(t, fromDir, srcTotal, "Src Dir")
	if err != nil {
		return err
	}

	dstTotal := testDstTotal(tdSrc, tdDst, dup)

	err = countFiles(t, toDir, dstTotal, "Target Dir")
	if err != nil {
		return err
	}

	return nil
}

func testMerge(t *testing.T, method Method, action Action, dstFileNo int, dup bool) error {
	tdSrc := newTestDir(t, method, fileNoDefault)
	tdDst := newTestDir(t, method, dstFileNo)

	src := tdSrc.buildRoot()
	dst := tdDst.buildRoot()

	// Copy files to two sorted directories that are identical
	fromDir := tdSrc.buildSortedDir(src, "fromDir_", ActionCopy)
	toDir := tdDst.buildSortedDir(dst, "toDir_", ActionCopy)

	// merge them
	m := NewMerger(fromDir, toDir, action, "")

	err := m.Merge(ioutil.Discard)
	if err != nil {
		return err
	}

	err = testMergeResults(t, tdSrc, tdDst, fromDir, toDir, action, dup)
	if err != nil {
		return err
	}

	defer os.RemoveAll(fromDir)
	defer os.RemoveAll(toDir)
	defer os.RemoveAll(dst)
	defer os.RemoveAll(src)

	return nil
}

func testMergeCollisions(t *testing.T, method Method, action Action) error {
	tdSrc := newTestDir(t, method, fileNoDefault)
	// Not for collisons the dst dir has to have the default fileNo
	tdDst := newTestDir(t, method, fileNoDefault)

	src := tdSrc.buildCollisionWithRoot()
	dst := tdDst.buildRoot()

	// Copy files to two sorted directories that are identical
	fromDir := tdSrc.buildSortedDir(src, "fromDir_", ActionCopy)
	toDir := tdDst.buildSortedDir(dst, "toDir_", ActionCopy)

	// merge them
	m := NewMerger(fromDir, toDir, action, "")

	err := m.Merge(ioutil.Discard)
	if err != nil {
		return err
	}

	err = testMergeResults(t, tdSrc, tdDst, fromDir, toDir, action, false)
	if err != nil {
		return err
	}

	defer os.RemoveAll(fromDir)
	defer os.RemoveAll(toDir)
	defer os.RemoveAll(dst)
	defer os.RemoveAll(src)

	return nil
}

func testMergeTimeSpread(t *testing.T, method Method, action Action) error {
	tdSrc := newTestDir(t, method, fileNoDefault)
	// Time should trump whether or not the name is the same
	tdDst := newTestDir(t, method, fileNoDefault)

	src := tdSrc.buildTimeSpreadRoot()
	dst := tdDst.buildRoot()

	// Copy files to two sorted directories that are identical
	fromDir := tdSrc.buildSortedDir(src, "fromDir_", ActionCopy)
	toDir := tdDst.buildSortedDir(dst, "toDir_", ActionCopy)

	// merge them
	m := NewMerger(fromDir, toDir, action, "")

	err := m.Merge(ioutil.Discard)
	if err != nil {
		return err
	}

	err = testMergeResults(t, tdSrc, tdDst, fromDir, toDir, action, false)
	if err != nil {
		return err
	}

	defer os.RemoveAll(fromDir)
	defer os.RemoveAll(toDir)
	defer os.RemoveAll(dst)
	defer os.RemoveAll(src)

	return nil
}

func testMergeFilter(t *testing.T, method Method, action Action) error {
	tdSrc := newTestDir(t, method, fileNoDefault)
	// Not for collisons the dst dir has to have the default fileNo
	tdDst := newTestDir(t, method, fileNoDefault)

	src := tdSrc.buildTifRoot()
	dst := tdDst.buildRoot()

	// Copy files to two sorted directories that are identical
	fromDir := tdSrc.buildSortedDir(src, "fromDir_", ActionCopy)
	toDir := tdDst.buildSortedDir(dst, "toDir_", ActionCopy)

	// merge but only transfer tif files
	regex := tdSrc.getTifRegex()

	m := NewMerger(fromDir, toDir, action, regex)

	err := m.Merge(ioutil.Discard)
	if err != nil {
		return err
	}

	var srcTotal int

	switch action {
	case ActionCopy:
		// src dir should have all its media untouched
		srcTotal = tdSrc.numTotal()
	case ActionMove:
		// src dir should have all media merged and be empty
		srcTotal = tdSrc.numTotal() - tdSrc.numTif
	default:
		return errors.New("unknown action")
	}

	err = countFiles(t, fromDir, srcTotal, "Src Dir")
	if err != nil {
		return err
	}

	dstTotal := tdDst.numData + tdSrc.numTif

	err = countFiles(t, toDir, dstTotal, "Target Dir")
	if err != nil {
		return err
	}

	defer os.RemoveAll(fromDir)
	defer os.RemoveAll(toDir)
	defer os.RemoveAll(dst)
	defer os.RemoveAll(src)

	return nil
}

func TestMergeGood(t *testing.T) {
	// By setting the fileNo so high the files will have different names
	// between tesdirs We are hoping that this number is just high enough
	// but as of this writing testdir has 150 files, and our countfiles
	// checking should protect us.
	fileNo := 10000

	for _, action := range Actions() {
		for _, method := range Methods() {
			action := action
			method := method
			name := action.String() + "/" + method.String()

			t.Run(name, func(t *testing.T) {
				t.Parallel()

				err := testMerge(t, method, action, fileNo, false)
				if err != nil {
					t.Fatalf("Method %s, Action %s Error: %s\n",
						action, method, err.Error())
				}
			})
		}
	}
}

func TestMergeTime(t *testing.T) {
	for _, action := range Actions() {
		for _, method := range Methods() {
			action := action
			method := method
			name := action.String() + "/" + method.String()

			t.Run(name, func(t *testing.T) {
				t.Parallel()
				err := testMergeTimeSpread(t, method, action)
				if err != nil {
					t.Fatalf("Method %s, Action Copy Error: %s\n",
						method, err.Error())
				}
			})
		}
	}
}

func TestMergeDuplicate(t *testing.T) {
	// By setting the fileNo to the default we ensure the dst directory will have
	// files with the same names as src and then get duplicates.
	fileNo := fileNoDefault

	for _, action := range Actions() {
		for _, method := range Methods() {
			action := action
			method := method
			name := action.String() + "/" + method.String()

			t.Run(name, func(t *testing.T) {
				t.Parallel()

				err := testMerge(t, method, action, fileNo, true)
				if err != nil {
					t.Fatalf("Method %s, Action Copy Error: %s\n",
						method, err.Error())
				}
			})
		}
	}
}

func TestMergeCollisions(t *testing.T) {
	for _, action := range Actions() {
		for _, method := range Methods() {
			action := action
			method := method
			name := action.String() + "/" + method.String()

			t.Run(name, func(t *testing.T) {
				t.Parallel()

				err := testMergeCollisions(t, method, action)
				if err != nil {
					t.Fatalf("Method %s, Action Copy Error: %s\n",
						method, err.Error())
				}
			})
		}
	}
}

func TestMergeMethodDiff(t *testing.T) {
	t.Parallel()
	tdSrc := newTestDir(t, MethodYear, fileNoDefault)
	tdDst := newTestDir(t, MethodMonth, fileNoDefault)

	src := tdSrc.buildRoot()
	dst := tdDst.buildRoot()

	// Copy files to two sorted directories that are identical
	fromDir := tdSrc.buildSortedDir(src, "fromDir_", ActionCopy)
	toDir := tdDst.buildSortedDir(dst, "toDir_", ActionCopy)

	m := NewMerger(fromDir, toDir, ActionCopy, "")

	err := m.Merge(ioutil.Discard)
	if err == nil {
		t.Errorf("Succes is unexpected. src and dst have the different methods\n")
	}
}

// Here we create a situation where one sorted driectory by a method has other
// artifacts of other methods in side. Not valid so we should detext and fail.
func TestMergeMethodMultiple(t *testing.T) {
	t.Parallel()
	tdSrc := newTestDir(t, MethodMonth, fileNoDefault)
	tdDst := newTestDir(t, MethodMonth, fileNoDefault)

	src := tdSrc.buildRoot()
	dst := tdDst.buildRoot()

	// Copy files to two sorted directories that are identical
	fromDir := tdSrc.buildSortedDir(src, "fromDir_", ActionCopy)
	toDir := tdDst.buildSortedDir(dst, "toDir_", ActionCopy)

	// Time to add some Year structure to a month directory.
	badDir := filepath.Join(toDir, "/2020/")
	_ = os.MkdirAll(badDir, 0700)
	badPath := filepath.Join(badDir, "badFile.jpg")
	_ = copyFile(exifPath, badPath)

	m := NewMerger(fromDir, toDir, ActionCopy, "")

	err := m.Merge(ioutil.Discard)
	if err == nil {
		t.Errorf("Succes is unexpected. src and dst have the different methods\n")
	}
}

func TestMergeFilter(t *testing.T) {
	for _, method := range Methods() {
		method := method

		t.Run(method.String(), func(t *testing.T) {
			t.Parallel()

			err := testMergeFilter(t, method, ActionCopy)
			if err != nil {
				t.Fatalf("Method %s, Action Copy Error: %s\n",
					method, err.Error())
			}
		})
	}

	for _, method := range Methods() {
		method := method

		t.Run(method.String(), func(t *testing.T) {
			t.Parallel()

			err := testMergeFilter(t, method, ActionMove)
			if err != nil {
				t.Fatalf("Method %s, Action Move Error: %s\n",
					method, err.Error())
			}
		})
	}
}
