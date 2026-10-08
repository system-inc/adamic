package native

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"os"
	"path/filepath"
)

// Ship the real compiler libraries in tsc's executable-relative layout.
func installStartupLibraries(directory string) error {
	fs := bundled.WrapFS(osvfs.FS())
	for _, name := range bundled.LibNames {
		source, ok := fs.ReadFile(bundled.LibPath().ResolveFile(name))
		if !ok {
			return fmt.Errorf("native: missing bundled TypeScript library %s", name)
		}
		destination := filepath.Join(directory, name)
		if existing, err := os.ReadFile(destination); err == nil {
			if string(existing) != source {
				return fmt.Errorf("native: executable library already exists with different contents: %s", destination)
			}
			continue
		} else if !os.IsNotExist(err) {
			return err
		}
		file, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			return err
		}
		_, err = file.WriteString(source)
		closeError := file.Close()
		if err != nil {
			return err
		}
		if closeError != nil {
			return closeError
		}
	}
	return nil
}
