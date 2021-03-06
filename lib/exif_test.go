package exifsort

import (
	"strings"
	"testing"
	"time"
)

func TestExifFormatError(t *testing.T) {
	testErrStr := "bad format for dingle: dangle Problem"
	err := newExifError("dangle", "dingle")

	if err.Error() != testErrStr {
		t.Errorf("Errors do not match: %s %s", err, testErrStr)
	}
}

func TestExifGoodTimes(t *testing.T) {
	good1String := "2008:03:01 12:36:01"
	good2String := "2008:03:01 12:36:01.34"
	testMonth := 3
	goodTime := time.Date(2008, time.Month(testMonth), 1, 12, 36, 1, 0, time.Local)

	testTime, err := extractTimeFromStr(good1String)
	if testTime != goodTime {
		t.Errorf("Return Time is incorrect %q\n", testTime)
	}

	if err != nil {
		t.Errorf("Error is incorrectly not nil %q\n", err)
	}

	testTime, err = extractTimeFromStr(good2String)
	if testTime != goodTime {
		t.Errorf("Return Time is incorrect %q\n", testTime)
	}

	if err != nil {
		t.Errorf("Error is incorrectly not nil %q\n", err)
	}
}

func TestExifExtractBadTimeFromStr(t *testing.T) {
	var formBadInput = map[string]string{
		"Gobo":                    "Space Problem",
		"Gobo a a a a":            "Space Problem",
		"Gobo Hey":                "Date Split",
		"Gobo:03:01 12:36:11":     "Year",
		"2008:Gobo:01 12:36:11":   "Month",
		"2008:03:Gobo 12:36:11":   "Day",
		"2008:03:01 Gobo":         "Time Split",
		"2008:03:01 Gobo:36:11":   "Hour",
		"2008:03:01 12:Gobo:11":   "Minute",
		"2008:03:01 12:36:Gobo":   "Sec",
		"2008:03:01 12:36:Gobo.2": "Sec",
	}

	for input, errLabel := range formBadInput {
		_, err := extractTimeFromStr(input)
		if err == nil {
			t.Fatalf("Expected error on input: %s\n", input)
		}

		if strings.Contains(err.Error(), errLabel) == false {
			t.Errorf("Improper error reporting on input %s: %s\n",
				input, err.Error())
		}
	}
}

func TestExifTimeErr(t *testing.T) {
	var testInput = map[string]bool{
		exifPath:       true,
		noExifPath:     false,
		noRootExifPath: false,
		nonesensePath:  false,
	}

	for path, valid := range testInput {
		path := path
		valid := valid

		t.Run(path, func(t *testing.T) {
			t.Parallel()
			_, err := ExifTimeGet(path)
			if valid == true && err != nil {
				t.Errorf("Unexpected Error with good input file %s\n", path)
			}

			if !valid && err == nil {
				t.Errorf("Expected error with invalid Exif file %s.\n", path)
			}
		})
	}
}

func TestExifTimeVal(t *testing.T) {
	t.Parallel()

	goodTime, _ := extractTimeFromStr(exifTimeStr)

	time, err := ExifTimeGet(exifPath)
	if err != nil {
		t.Errorf("Unexpected Error with good input file %s\n", exifPath)
	}

	if goodTime != time {
		t.Errorf("Expected Time %s but got %s\n", goodTime, time)
	}
}
