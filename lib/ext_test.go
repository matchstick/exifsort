package exifsort

import (
	"fmt"
	"strings"
	"testing"
)

func TestSkipFileType(t *testing.T) {
	// Try just gobo.<extension>
	for extension := range extensionMap() {
		goodInput := fmt.Sprintf("gobo.%s", extension)

		skip := skipFileType(goodInput)
		if skip {
			t.Errorf("Expected False for %s\n", goodInput)
		}
	}
	// Try a simple upper case just gobo.<extension>
	for extension := range extensionMap() {
		goodInput := strings.ToUpper(fmt.Sprintf("gobo.%s", extension))

		skip := skipFileType(goodInput)
		if skip {
			t.Errorf("Expected False for %s\n", goodInput)
		}
	}

	// Try with many "." hey.gobo.<extension>
	for extension := range extensionMap() {
		goodInput := fmt.Sprintf("hey.gobo.%s", extension)

		skip := skipFileType(goodInput)
		if skip {
			t.Errorf("Expected False for %s\n", goodInput)
		}
	}

	badInput := "gobobob.."

	skip := skipFileType(badInput)
	if !skip {
		t.Errorf("Expected True for %s\n", badInput)
	}

	badInput = "gobo"

	skip = skipFileType(badInput)
	if !skip {
		t.Errorf("Expected True for %s\n", badInput)
	}

	// Try ".." at the end.<extension>
	for extension := range extensionMap() {
		badInput := fmt.Sprintf("gobo.%s..", extension)

		skip := skipFileType(badInput)
		if !skip {
			t.Errorf("Expected True for %s\n", badInput)
		}
	}
}

func TestSkipSynologyTypes(t *testing.T) {
	badInput := "@eaDir"

	skip := skipFileType(badInput)
	if !skip {
		t.Errorf("Expected True for %s\n", badInput)
	}

	badInput = "@syno"

	skip = skipFileType(badInput)
	if !skip {
		t.Errorf("Expected True for %s\n", badInput)
	}

	badInput = "synofile_thumb"

	skip = skipFileType(badInput)
	if !skip {
		t.Errorf("Expected True for %s\n", badInput)
	}
}
