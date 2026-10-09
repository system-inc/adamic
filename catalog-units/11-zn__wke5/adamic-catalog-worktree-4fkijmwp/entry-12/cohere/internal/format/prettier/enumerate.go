package prettier

import "github.com/system-inc/cohere/internal/format/formatfiles"

// Enumerate walks a project root and returns the files this engine would format.
func (engine *Engine) Enumerate(root string) (formatfiles.Enumeration, error) {
	return formatfiles.Enumerate(root, engine.Handles)
}
