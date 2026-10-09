//go:build !darwin && !linux

package program

// mapFile maps nothing here, which leaves the content pack empty.
func mapFile(path string) []byte {
	return nil
}
