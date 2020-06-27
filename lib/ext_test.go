package exifsort

import (
	"fmt"
	"strings"
	"testing"
)

func TestExtSkipFileType(t *testing.T) {
	// Try just gobo.<extension>
	for _, extension := range ExtensionsPhoto() {
		goodInput := fmt.Sprintf("gobo.%s", extension)

		category := categorizeFile(goodInput)
		if category == categorySkip {
			t.Errorf("Expected to not skip for %s\n", goodInput)
		}
	}
	// Try a simple upper case just gobo.<extension>
	for _, extension := range ExtensionsPhoto() {
		goodInput := strings.ToUpper(fmt.Sprintf("gobo.%s", extension))

		category := categorizeFile(goodInput)
		if category == categorySkip {
			t.Errorf("Expected to not skip for %s\n", goodInput)
		}
	}

	// Try with many "." hey.gobo.<extension>
	for _, extension := range ExtensionsPhoto() {
		goodInput := fmt.Sprintf("hey.gobo.%s", extension)

		category := categorizeFile(goodInput)
		if category == categorySkip {
			t.Errorf("Expected to not skip for %s\n", goodInput)
		}
	}

	badInput := "gobobob.."

	category := categorizeFile(badInput)
	if category != categorySkip {
		t.Errorf("Expected to skip for %s\n", badInput)
	}

	badInput = "gobo"

	category = categorizeFile(badInput)
	if category != categorySkip {
		t.Errorf("Expected to skip for %s\n", badInput)
	}
}

func TestExtSkipSynologyTypes(t *testing.T) {
	badInput := []string{
		".DS_Store@SynoResource",
		"2019@SynoEAStream",
		"PSD_Work/@eaDir/",
		"@eaDir",
		"@syno",
		"IMG_0269.JPG/SYNOFILE_THUMB_M_r1.jpg",
	}

	for _, input := range badInput {
		category := categorizeFile(input)
		if category != categorySkip {
			t.Errorf("Expected to skip for %s\n", input)
		}
	}
}
