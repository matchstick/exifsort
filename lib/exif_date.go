package exifSort

import (
	"fmt"
	"os"

	"io/ioutil"

	"github.com/dsoprea/go-logging"

	"github.com/dsoprea/go-exif/v2"
	"github.com/dsoprea/go-exif/v2/common"
)

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
	Path string
	Date string
}

func ExtractExifDate(filepath string) (entry ExifDateEntry, err error) {
	var exifDateEntry = ExifDateEntry { false, filepath, "" }

	f, err := os.Open(filepath)
	log.PanicIf(err)

	data, err := ioutil.ReadAll(f)
	log.PanicIf(err)

	rawExif, err := exif.SearchAndExtractExif(data)
	if err != nil {
		if err == exif.ErrNoExif {
			fmt.Printf("No EXIF data.\n")
			return exifDateEntry, nil
		}

		log.Panic(err)
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
				// Decide to turn into a log error or not
//				fmt.Printf("WARNING: Unknown tag: [%s] (%04x)\n", ifdPath, tagId)
				return nil
			} else {
				log.Panic(err)
			}
		}

		value, err := ite.Value()
		if err != nil {
			if log.Is(err, exifcommon.ErrUnhandledUndefinedTypedTag) == true {
				fmt.Printf("WARNING: Non-standard undefined tag: [%s] (%04x)\n", ifdPath, tagId)
				return nil
			}

			log.Panic(err)
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
	log.PanicIf(err)

	for _, entry := range entries {
		// TODO Is this the best field? from quick googling it looks
		// like the most reliable.
		if entry.TagName == "DateTimeOriginal" {
			//TODO figure out time
			exifDateEntry.Date = entry.ValueString
		}
	}
	
	exifDateEntry.Valid = true
	return exifDateEntry, nil
}
