package native

import (
	"fmt"
	"path/filepath"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/format/formatfiles"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/gitignore"
)

// Resolving formats each file with the options its own directory resolves to, the way
// prettier.Resolving does for the goja engine, so switching the format phase from one to the other
// changes the printers and nothing about which options a file is printed with.
//
// It holds no engines, because a native printer costs nothing to construct: only the run's Resolver,
// which remembers each directory's options.
type Resolving struct {
	resolver *formatoptions.Resolver
}

// NewResolving resolves startDirectory's options eagerly, so a config cohere refuses stops the run
// before any phase does work rather than on the first file after fixes have begun.
func NewResolving(startDirectory string) (*Resolving, error) {
	resolving := &Resolving{resolver: formatoptions.NewResolver()}
	if _, err := resolving.resolver.Resolve(startDirectory); err != nil {
		return nil, err
	}
	return resolving, nil
}

// Handles reports whether a native printer is registered for the file's type.
func (resolving *Resolving) Handles(fileName string) bool {
	_, present := lookup(fileName)
	return present
}

// Enumerate walks a tree for the files a native printer handles.
func (resolving *Resolving) Enumerate(root string) (formatfiles.Enumeration, error) {
	return formatfiles.Enumerate(root, resolving.Handles)
}

// EnumerateListed is Enumerate reading directories from listing where it holds them. See
// formatfiles.EnumerateListed.
func (resolving *Resolving) EnumerateListed(root string, listing gitignore.Listing) (formatfiles.Enumeration, error) {
	return formatfiles.EnumerateListed(root, resolving.Handles, listing)
}

// Format formats one file with the options its own directory resolves to.
func (resolving *Resolving) Format(fileName string, text string) (string, error) {
	return resolving.FormatParsed(fileName, text, nil)
}

// FormatParsed is Format with a tree of the text someone already parsed, or nil (Formatter.FormatParsed).
func (resolving *Resolving) FormatParsed(fileName string, text string, parsed *ast.SourceFile) (string, error) {
	resolution, err := resolving.resolver.Resolve(filepath.Dir(fileName))
	if err != nil {
		return "", fmt.Errorf("resolving the format options for %s: %w", fileName, err)
	}
	return Formatter{Options: resolution.Options}.FormatParsed(fileName, text, parsed)
}

// OptionsFingerprint names the options a file formats with: every field, by name, so two resolutions
// that differ anywhere print differently.
func (resolving *Resolving) OptionsFingerprint(fileName string) (string, error) {
	resolution, err := resolving.resolver.Resolve(filepath.Dir(fileName))
	if err != nil {
		return "", fmt.Errorf("resolving the format options for %s: %w", fileName, err)
	}
	return fmt.Sprintf("%+v", resolution.Options), nil
}
