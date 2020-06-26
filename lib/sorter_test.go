package exifsort

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"
)

func testTransfer(t *testing.T, td *testdir, method Method, action Action) error {
	scanner := NewScanner()
	_ = scanner.ScanDir(td.root, ioutil.Discard)

	dst, _ := ioutil.TempDir("", "sort_dst_")
	defer os.RemoveAll(dst)

	sorter, err := NewSorter(scanner, method)
	if err != nil {
		return err
	}

	err = sorter.Transfer(dst, action, ioutil.Discard)
	if err != nil {
		return err
	}

	err = countFiles(t, dst, td.numData, "Dst Data")
	if err != nil {
		return err
	}

	switch {
	case action == ActionCopy:
		err := countFiles(t, td.root, td.numTotal(), "Src Copy")
		if err != nil {
			return err
		}

	case action == ActionMove:
		leftovers := td.numScanError + td.numSkipped

		err := countFiles(t, td.root, leftovers, "Src Move")
		if err != nil {
			return err
		}
	default:
		return errors.New("unknown action")
	}

	return nil
}

func TestSortDir(t *testing.T) {
	t.Parallel()
	for _, method := range Methods() {
		td := newTestDir(t, method, fileNoDefault)

		src := td.buildRoot()
		defer os.RemoveAll(src)

		err := testTransfer(t, td, method, ActionCopy)
		if err != nil {
			t.Errorf("%s\n", err.Error())
		}

		err = testTransfer(t, td, method, ActionMove)
		if err != nil {
			t.Errorf("%s\n", err.Error())
		}
	}
}

func TestSortDuplicates(t *testing.T) {
	t.Parallel()
	for _, method := range Methods() {
		td := newTestDir(t, method, fileNoDefault)

		src := td.buildDuplicateWithinThisRoot()
		defer os.RemoveAll(src)

		err := testTransfer(t, td, method, ActionCopy)
		if err != nil {
			t.Errorf("%s\n", err.Error())
		}

		err = testTransfer(t, td, method, ActionMove)
		if err != nil {
			t.Errorf("%s\n", err.Error())
		}
	}
}

func TestSortCollisions(t *testing.T) {
	t.Parallel()
	for _, method := range Methods() {
		td := newTestDir(t, method, fileNoDefault)

		src := td.buildCollisionWithinThisRoot()
		defer os.RemoveAll(src)

		err := testTransfer(t, td, method, ActionCopy)
		if err != nil {
			t.Errorf("%s\n", err.Error())
		}

		err = testTransfer(t, td, method, ActionMove)
		if err != nil {
			t.Errorf("%s\n", err.Error())
		}
	}
}

func TestBadSortMethod(t *testing.T) {
	t.Parallel()
	const badMethod = 888

	td := newTestDir(t, badMethod, fileNoDefault)

	err := testTransfer(t, td, badMethod, ActionMove)
	if err == nil {
		t.Fatalf("Expected error got success\n")
	}
}

func TestBadSortAction(t *testing.T) {
	t.Parallel()
	const badAction = 888

	td := newTestDir(t, MethodNone, fileNoDefault)

	err := testTransfer(t, td, MethodYear, badAction)
	if err == nil {
		t.Fatalf("Expected error got success\n")
	}
}

func TestSortNoOutputDir(t *testing.T) {
	t.Parallel()
	td := newTestDir(t, MethodNone, fileNoDefault)

	src := td.buildRoot()
	defer os.RemoveAll(src)

	scanner := NewScanner()
	_ = scanner.ScanDir(src, ioutil.Discard)

	sorter, _ := NewSorter(scanner, MethodYear)

	err := sorter.Transfer("dst", ActionMove, ioutil.Discard)
	if err == nil {
		t.Fatalf("Unexpected Success\n")
	}
}
