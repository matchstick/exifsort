package exifSort

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestFormatError(t *testing.T) {
	testErr := errors.New("Bad Format for dingle: dangle Problem\n")
	_, err := formatError("dangle", "dingle")
	if err.Error() != testErr.Error() {
		t.Errorf("Errors do not match: %s %s", err, testErr)
	}
}

type formError struct {
	input    string
	errLabel string
}

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

func TestGoodTimes(t *testing.T) {
	good1String := "2008:03:01 12:36:01"
	good2String := "2008:03:01 12:36:01.34"
	goodTime := time.Date(2008, time.Month(3), 1, 12, 36, 1, 0, time.Local)

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

func TestExtractExifTime(t *testing.T) {

	validExifPath := "../data/with_exif.jpg"
	goodDateStr := "2020:04:28 14:12:21"
	goodTime, _ := extractTimeFromStr(goodDateStr)

	time, err := ExtractExifTime(validExifPath)
	if err != nil {
		t.Errorf("Unexpected Error with good input file\n")
	}

	if goodTime != time {
		t.Errorf("Expected Time %s but got %s\n", goodTime, time)
	}

	invalidExifPath := "../data/no_exif.jpg"

	time, err = ExtractExifTime(invalidExifPath)
	if err == nil {
		t.Errorf("Unexpected success with invalid Exif file.\n")
	}

	nonePath := "../gobofragggle"
	time, err = ExtractExifTime(nonePath)
	if err == nil {
		t.Errorf("Unexpected success with nonsense path\n")
	}

}
