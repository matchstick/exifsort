package exifSort

import (
	"fmt"
	"github.com/dsoprea/go-exif-knife"
	"github.com/dsoprea/go-exif/v2"
	"strconv"
	"strings"
	"time"
)

func formatError(label string, dateString string) (time.Time, error) {
	var t time.Time
	return t, fmt.Errorf("Bad Format for %s: %s Problem\n", dateString, label)
}

// Seconds are funny. The format may be "<sec> <milli>"
// or it may be with an extra decmial place such as <sec>.<hundredths>
func extractSecsFractionFromStr(secsStr string) (int, error) {
	splitSecs := strings.Split(secsStr, ".")
	if len(splitSecs) != 2 {
		return 0, fmt.Errorf("Not a fraction second")
	}

	// We only care about what is in front of the "."
	secs, err := strconv.Atoi(splitSecs[0])
	if err != nil {
		return 0, fmt.Errorf("Not a convertaible second")
	}
	return secs, nil
}

func extractTimeFromStr(exifDateTime string) (time.Time, error) {
	splitDateTime := strings.Split(exifDateTime, " ")
	if len(splitDateTime) != 2 {
		return formatError("Space Problem", exifDateTime)
	}
	date := splitDateTime[0]
	timeOfDay := splitDateTime[1]

	splitDate := strings.Split(date, ":")
	if len(splitDate) != 3 {
		return formatError("Date Split", exifDateTime)
	}

	year, err := strconv.Atoi(splitDate[0])
	if err != nil {
		return formatError("Year", exifDateTime)
	}

	month, err := strconv.Atoi(splitDate[1])
	if err != nil {
		return formatError("Month", exifDateTime)
	}

	day, err := strconv.Atoi(splitDate[2])
	if err != nil {
		return formatError("Day", exifDateTime)
	}

	splitTime := strings.Split(timeOfDay, ":")
	if len(splitTime) != 3 {
		return formatError("Time Split", exifDateTime)
	}

	hour, err := strconv.Atoi(splitTime[0])
	if err != nil {
		return formatError("Hour", exifDateTime)
	}

	minute, err := strconv.Atoi(splitTime[1])
	if err != nil {
		return formatError("Minute", exifDateTime)
	}

	second, err := strconv.Atoi(splitTime[2])
	if err != nil {
		second, err = extractSecsFractionFromStr(splitTime[2])
		if err != nil {
			return formatError("Sec", exifDateTime)
		}
	}

	t := time.Date(year, time.Month(month), day,
		hour, minute, second, 0, time.Local)
	return t, nil
}

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
		return time, fmt.Errorf("Root Ifd not found.")
	}

	// See if the EXIF info path is there. We want DateTimeOriginal
	exifIfd, err := exif.FindIfdFromRootIfd(mc.RootIfd, "IFD/Exif")
	if err != nil {
		return time, fmt.Errorf("IFD/Exif not found.")
	}

	// Query for DateTimeOriginal
	results, err := exifIfd.FindTagWithName("DateTimeOriginal")
	if err != nil {
		return time, fmt.Errorf("DateTimeOriginal Tag not found.")
	}
	if len(results) != 1 {
		return time, fmt.Errorf("Too many DateTimeOriginal Tags found.")
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
