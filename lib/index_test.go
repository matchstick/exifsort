package exifSort

// TODO this test file is a mess. But it found many bugs.
// Now that we know what we want we will rewrite.

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"
)

var ByYearInput = []string{
	"IMG.jpg",
	"c.jpg",
	"c.jpg",
	"a.jpg",
	"c.jpg",
}

const exifFile = "../data/with_exif.jpg"
const diffFile = "../data/diff_exif.jpg"
const diff2File = "../data/diff2_exif.jpg"

func indexTmpDir(t *testing.T, parent string, name string) string {
	newDir, err := ioutil.TempDir(parent, name)
	if err != nil {
		t.Fatal(err)
	}
	return newDir
}

func copyFile(t *testing.T, src string, dst string) {
	content, err := ioutil.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(dst, content, 0644)
	if err != nil {
		t.Fatal(err)
	}
}

// Add Some filepaths to the index
// File names are of the form "IMG_<start>.jpg" to "IMG_<end>.jpg"
// We have a time associated with each file based on args provided.
func PutFiles(t *testing.T, idx index, dir string, srcFile string,
	start uint, count uint,
	year int, month int, day int) {
	for ii := start; ii < start+count; ii++ {
		time := time.Date(year, time.Month(month), day,
			0, 0, 0, 0, time.Local)
		name := fmt.Sprintf("%s/IMG_%d.jpg", dir, ii)
		copyFile(t, srcFile, name)
		err := idx.Put(name, time)
		if err != nil {
			t.Error(err)
		}
	}
}

func GetFiles(t *testing.T, idx index, start uint, count uint,
	year int, month int, day int) {
	for ii := start; ii < start+count; ii++ {
		testTime := time.Date(year, time.Month(month), day,
			0, 0, 0, 0, time.Local)
		name := fmt.Sprintf("IMG_%d.jpg", ii)
		idxPath, present := idx.Get(name)
		if present == false {
			t.Fatalf("Cannot find %s\n", name)
		}
		testPath := idx.PathStr(testTime, name)
		if idxPath != testPath {
			t.Fatalf("Recvd paths (%s) != %s using time %s\n",
				idxPath, testPath, testTime)
		}
	}
}

func GetCollisionFiles(t *testing.T, idx index, start uint, count uint,
	year int, month int, day int, collisionCount int) {
	for ii := start; ii < start+count; ii++ {
		testTime := time.Date(year, time.Month(month), day,
			0, 0, 0, 0, time.Local)
		name := fmt.Sprintf("IMG_%d_%d.jpg", ii, collisionCount)
		idxPath, present := idx.Get(name)
		if present == false {
			t.Fatalf("Cannot find %s\n", name)
		}
		testPath := idx.PathStr(testTime, name)
		if idxPath != testPath {
			t.Fatalf("Recvd collison paths (%s) != %s using time %s\n",
				idxPath, testPath, testTime)
		}
	}
}
func indexSizeCheck(t *testing.T, targetSize int, idx index) {
	idxMap := idx.GetAll()
	mapSize := len(idxMap)
	if mapSize != targetSize {
		t.Errorf("Expecting to have index hold %d entries not %d\n", targetSize, mapSize)
	}
}

func TestIndexPutGet(t *testing.T) {
	for method := METHOD_YEAR; method < METHOD_NONE; method++ {
		var idx = createIndex(METHOD_YEAR)
		testDir := indexTmpDir(t, "", "root")

		PutFiles(t, idx, testDir, exifFile, 10, 10, 2020, 1, 1)
		GetFiles(t, idx, 10, 10, 2020, 1, 1)
		indexSizeCheck(t, 10, idx)
	}
}

func TestIndexCollisions(t *testing.T) {
	for method := METHOD_YEAR; method < METHOD_NONE; method++ {
		var idx = createIndex(method)
		testDir := indexTmpDir(t, "", "root_")
		testDir1 := indexTmpDir(t, "", "bobo_")
		testDir2 := indexTmpDir(t, "", "gobo_")
		PutFiles(t, idx, testDir, exifFile, 10, 10, 1, 2, 4)
		PutFiles(t, idx, testDir1, diffFile, 10, 10, 1, 2, 4)
		PutFiles(t, idx, testDir2, diff2File, 10, 10, 1, 2, 4)
		GetCollisionFiles(t, idx, 10, 10, 1, 2, 4, 0)
		GetCollisionFiles(t, idx, 10, 10, 1, 2, 4, 1)
		indexSizeCheck(t, 30, idx)
	}
}

func TestIndexDuplicates(t *testing.T) {
	for method := METHOD_YEAR; method < METHOD_NONE; method++ {
		var exifPath = "../data/with_exif.jpg"
		var idx = createIndex(METHOD_YEAR)
		time, _ := ExtractExifTime(exifPath)
		idx.Put(exifPath, time)
		err := idx.Put(exifPath, time)
		if err == nil {
			t.Error("Expected Error with duplicate Put. Got nil\n")
		}
		indexSizeCheck(t, 1, idx)
	}
}
