package exifsort

import (
	"strings"
	"testing"
	"time"
)

func TestFormatError(t *testing.T) {
	testErrStr := "bad format for dingle: dangle Problem"
	err := newExifError("dangle", "dingle")

	if err.Error() != testErrStr {
		t.Errorf("Errors do not match: %s %s", err, testErrStr)
	}
}

func TestGoodTimes(t *testing.T) {
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

func TestExtractBadTimeFromStr(t *testing.T) {
	var formBadInput = map[string]string{
		"Gobo":                  "Space Problem",
		"Gobo a a a a":          "Space Problem",
		"Gobo Hey":              "Date Split",
		"Gobo:03:01 12:36:11":   "Year",
		"2008:Gobo:01 12:36:11": "Month",
		"2008:03:Gobo 12:36:11": "Day",
		"2008:03:01 Gobo":       "Time Split",
		"2008:03:01 Gobo:36:11": "Hour",
		"2008:03:01 12:Gobo:11": "Minute",
		"2008:03:01 12:36:Gobo": "Sec",
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

const validExifPath = "../data/with_exif.jpg"
const invalidExifPath = "../data/no_exif.jpg"
const noRootExifPath = "../data/no_root_ifd.jpg"
const goodDateExifStr = "2020:04:28 14:12:21"
const goodDateStr = "2020/04/28 14:12:21"

func TestExtractTime(t *testing.T) {
	goodTime, _ := extractTimeFromStr(goodDateExifStr)

	time, err := ExtractTime(validExifPath)
	if err != nil {
		t.Errorf("Unexpected Error with good input file\n")
	}

	if goodTime != time {
		t.Errorf("Expected Time %s but got %s\n", goodTime, time)
	}

	_, err = ExtractTime(invalidExifPath)
	if err == nil {
		t.Errorf("Unexpected success with invalid Exif file.\n")
	}

	_, err = ExtractTime(noRootExifPath)
	if err == nil {
		t.Errorf("Unexpected success with invalid Exif file.\n")
	}

	nonePath := "../gobofragggle"

	_, err = ExtractTime(nonePath)
	if err == nil {
		t.Errorf("Unexpected success with nonsense path\n")
	}
}

func TestExtractTimeStr(t *testing.T) {
	str, err := ExtractTimeStr(validExifPath)
	if err != nil {
		t.Errorf("Unexpected Error with good input file\n")
	}

	if str != goodDateStr {
		t.Errorf("Expected String %s but got %s\n", goodDateStr, str)
	}

	_, err = ExtractTimeStr(invalidExifPath)
	if err == nil {
		t.Errorf("Unexpected success with invalid Exif file.\n")
	}
}
