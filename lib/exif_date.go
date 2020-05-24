package exifsort

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	exifknife "github.com/dsoprea/go-exif-knife"
	"github.com/dsoprea/go-exif/v2"
)

type exifError struct {
	prob string
}

func (e exifError) Error() string {
	return e.prob
}

func newExifError(label string, dateString string) error {
	var e exifError
	e.prob = fmt.Sprintf("bad format for %s: %s Problem", dateString, label)

	return e
}

const numSecsSplit = 2 // we expect two pieces

// Seconds are funny. The format may be "<sec> <milli>"
// or it may be with an extra decmial place such as <sec>.<hundredths>.
func secsFractionFromStr(secsStr string) (int, error) {
	splitSecs := strings.Split(secsStr, ".")
	if len(splitSecs) != numSecsSplit {
		return 0, &exifError{"Not a fraction second"}
	}

	// We only care about what is in front of the "."
	secs, err := strconv.Atoi(splitSecs[0])
	if err != nil {
		return 0, &exifError{"Not a convertaible second"}
	}

	return secs, nil
}

const numDateSplit = 3 // We expect the date to be X:X:X

func dateFromStr(str string, exifDateTime string) (int, int, int, error) {
	splitDate := strings.Split(str, ":")
	if len(splitDate) != numDateSplit {
		return 0, 0, 0, newExifError("Date Split", exifDateTime)
	}

	year, err := strconv.Atoi(splitDate[0])
	if err != nil {
		return 0, 0, 0, newExifError("Year", exifDateTime)
	}

	month, err := strconv.Atoi(splitDate[1])
	if err != nil {
		return 0, 0, 0, newExifError("Month", exifDateTime)
	}

	day, err := strconv.Atoi(splitDate[2])
	if err != nil {
		return 0, 0, 0, newExifError("Day", exifDateTime)
	}

	return year, month, day, nil
}

const numTimeSplit = 3 // We expect time to be X:X:X

func timeFromStr(str string, exifDateTime string) (int, int, int, error) {
	splitTime := strings.Split(str, ":")
	if len(splitTime) != numTimeSplit {
		return 0, 0, 0, newExifError("Time Split", exifDateTime)
	}

	hour, err := strconv.Atoi(splitTime[0])
	if err != nil {
		return 0, 0, 0, newExifError("Hour", exifDateTime)
	}

	minute, err := strconv.Atoi(splitTime[1])
	if err != nil {
		return 0, 0, 0, newExifError("Minute", exifDateTime)
	}

	second, err := strconv.Atoi(splitTime[2])
	if err != nil {
		second, err = secsFractionFromStr(splitTime[2])
		if err != nil {
			return 0, 0, 0, newExifError("Sec", exifDateTime)
		}
	}

	return hour, minute, second, nil
}

const numDateTimeSplit = 2 // We expect DateTime to be "Date Time"

func extractTimeFromStr(exifDateTime string) (time.Time, error) {
	var t time.Time

	splitDateTime := strings.Split(exifDateTime, " ")
	if len(splitDateTime) != numDateTimeSplit {
		return t, newExifError("Space Problem", exifDateTime)
	}

	date := splitDateTime[0]
	timeOfDay := splitDateTime[1]

	year, month, day, err := dateFromStr(date, exifDateTime)
	if err != nil {
		return t, err
	}

	hour, minute, second, err := timeFromStr(timeOfDay, exifDateTime)
	if err != nil {
		return t, err
	}

	t = time.Date(year, time.Month(month), day,
		hour, minute, second, 0, time.Local)

	return t, nil
}

const validDateTimeOrigninalTagNum = 1

// ExtractTime accepts a filepath, reads the exifdata stored inside and
// returns the 'Exif/DateTimeOriginal' value as a golang time.Time format.
//
// It returns an error if the file has no exif data, mangled exif data, or the
// contents are unexpected.
func ExtractTime(filepath string) (time.Time, error) {
	var time time.Time
	// Get the Exif Data and Ifd root
	mc, err := exifknife.GetExif(filepath)
	if err != nil {
		return time, err
	}
	// If the root is not there there is no exif data
	if mc.RootIfd == nil {
		return time, &exifError{"root ifd not found"}
	}

	// See if the EXIF info path is there. We want DateTimeOriginal
	exifIfd, err := exif.FindIfdFromRootIfd(mc.RootIfd, "IFD/Exif")
	if err != nil {
		return time, &exifError{"media IFD/Exif not found"}
	}

	// Query for DateTimeOriginal
	results, err := exifIfd.FindTagWithName("DateTimeOriginal")
	if err != nil {
		return time, &exifError{"the DateTimeOriginal Tag was not found"}
	}

	if len(results) != validDateTimeOrigninalTagNum {
		return time, &exifError{"too many DateTimeOriginal Tags found"}
	}

	// Found it, so extract value
	value, _ := results[0].Value()

	// Parse string into Time
	time, err = extractTimeFromStr(value.(string))
	if err != nil {
		return time, err
	}

	return time, nil
}

// ExtractTimeStr accepts a filepath, reads the exifdata stored inside and
// returns the 'Exif/DateTimeOriginal' value as a string.
//
// It returns an error if the file has no exif data, mangled exif data, or the
// contents are unexpected.
func ExtractTimeStr(path string) (string, error) {
	time, err := ExtractTime(path)
	if err != nil {
		return "", err
	}

	return exifTimeToStr(time), nil
}
