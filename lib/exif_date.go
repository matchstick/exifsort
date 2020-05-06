package exifSort

import (
	"fmt"
	"os"

	"io/ioutil"

	"github.com/dsoprea/go-logging"

	"github.com/dsoprea/go-exif/v2"
	"github.com/dsoprea/go-exif/v2/common"
	"github.com/dsoprea/go-exif/v2/undefined"

	"strconv"
	"strings"
	"time"
)

func formatError(label string, dateString string) (time.Time, error) {
	var t time.Time
	return t, fmt.Errorf("Bad Format for %s: %s Problem\n", dateString, label)
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
		return formatError("Sec", exifDateTime)
	}

	t := time.Date(year, time.Month(month), day,
		hour, minute, second, 0, time.Local)
	return t, nil
}

type IfdEntry struct {
	IfdPath     string                      `json:"ifd_path"`
	FqIfdPath   string                      `json:"fq_ifd_path"`
	IfdIndex    int                         `json:"ifd_index"`
	TagId       uint16                      `json:"tag_id"`
	TagName     string                      `json:"tag_name"`
	TagTypeId   exifcommon.TagTypePrimitive `json:"tag_type_id"`
	TagTypeName string                      `json:"tag_type_name"`
	UnitCount   uint32                      `json:"unit_count"`
	Value       interface{}                 `json:"value"`
	ValueString string                      `json:"value_string"`
}

type ExifDateEntry struct {
	Valid bool
	Path  string
	Time  time.Time
}

func ExtractExifDate(filepath string) (entry ExifDateEntry, err error) {
	var exifDateEntry ExifDateEntry
	exifDateEntry.Valid = false
	exifDateEntry.Path = filepath

	f, err := os.Open(filepath)
	if err != nil {
		return entry, err
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return entry, err
	}

	rawExif, err := exif.SearchAndExtractExif(data)
	if err != nil {
		if err == exif.ErrNoExif {
			return exifDateEntry, nil
		}
		return exifDateEntry, err
	}

	// Run the parse.
	im := exif.NewIfdMappingWithStandard()
	ti := exif.NewTagIndex()

	entries := make([]IfdEntry, 0)
	visitor := func(fqIfdPath string, ifdIndex int, ite *exif.IfdTagEntry) (err error) {
		defer func() {
			if state := recover(); state != nil {
				err = log.Wrap(state.(error))
				log.Panic(err)
			}
		}()

		tagId := ite.TagId()
		tagType := ite.TagType()

		ifdPath, err := im.StripPathPhraseIndices(fqIfdPath)
		log.PanicIf(err)

		it, err := ti.Get(ifdPath, tagId)
		if err != nil {
			if log.Is(err, exif.ErrTagNotFound) {
				return nil
			} else {
				return err
			}
		}

		value, err := ite.Value()
		if err != nil {
			if log.Is(err, exifcommon.ErrUnhandledUndefinedTypedTag) == true {
				return nil
			} else if err == exifundefined.ErrUnparseableValue {
				return nil
			}

			return err
		}

		valueString, err := ite.FormatFirst()
		log.PanicIf(err)

		entry := IfdEntry{
			IfdPath:     ifdPath,
			FqIfdPath:   fqIfdPath,
			IfdIndex:    ifdIndex,
			TagId:       tagId,
			TagName:     it.Name,
			TagTypeId:   tagType,
			TagTypeName: tagType.String(),
			UnitCount:   ite.UnitCount(),
			Value:       value,
			ValueString: valueString,
		}

		entries = append(entries, entry)

		return nil
	}

	_, err = exif.Visit(exifcommon.IfdStandard, im, ti, rawExif, visitor)
	if err != nil {
		return exifDateEntry, err
	}

	for _, entry := range entries {
		// TODO Is this the best field? from quick googling it looks
		// like the most reliable.
		if entry.TagName == "DateTimeOriginal" {
			exifDateEntry.Valid = true
			exifDateEntry.Time, err =
				extractTimeFromStr(entry.ValueString)
			if err != nil {
				return exifDateEntry, err
			}
			break
		}
	}

	return exifDateEntry, nil
}
