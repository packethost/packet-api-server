package util

import (
	"fmt"
	"strings"
)

// VolumeIDToName convert the uuid of a volume into the packet standard name
func VolumeIDToName(id string) string {
	// "3ee59355-a51a-42a8-b848-86626cc532f0" -> "volume-3ee59355"
	uuidElements := strings.Split(id, "-")
	return fmt.Sprintf("volume-%s", uuidElements[0])
}
