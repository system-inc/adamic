package micromark

import (
	"fmt"
	"strings"
)

func AdamicInputChunks(units []uint16) string {
	parts := []string{}
	for _, chunk := range preprocess(units) {
		if !chunk.IsText {
			parts = append(parts, fmt.Sprint("C", int(chunk.Code)))
			continue
		}
		text := []string{}
		for _, unit := range chunk.Text {
			text = append(text, fmt.Sprint(unit))
		}
		parts = append(parts, "T"+strings.Join(text, ","))
	}
	return strings.Join(parts, ";")
}
