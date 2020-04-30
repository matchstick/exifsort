package exifSort

import (
	"fmt"
	"time"
)

func ExifTime(t time.Time) string {
	return fmt.Sprintf("%d/%02d/%02d %02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
}
