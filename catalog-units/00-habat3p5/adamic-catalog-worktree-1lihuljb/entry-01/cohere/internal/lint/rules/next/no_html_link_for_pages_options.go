package next

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoHtmlLinkForPagesOptions is `@next/next/no-html-link-for-pages`'s one option element: the pages
// directory, or a list of them, that replace the default `pages/` and `src/pages/` under the root.
//
// Upstream's schema is `oneOf: [string, array of unique strings]`, read as
// `customPagesDirectory ? [customPagesDirectory] : defaults`, flattened. So the shapes mean:
//
//	absent or ""     the defaults, `pages/` and `src/pages/` under the root ("" is falsy upstream)
//	"dir"            that one directory
//	["a", "b"]       those directories
//	[]               no pages directory at all, since an empty array is truthy upstream
//
// Only the pages directories move. The app directories, `app/` and `src/app/`, are always looked for
// under the root, as upstream builds them from the root directories alone.
type NoHtmlLinkForPagesOptions struct {
	// Configured is whether the option replaced the default pages directories. False is absent or "".
	Configured bool

	// PagesDirectories are the configured directories, each absolute once decoded, or "" for an
	// element written as "", which is never found. Read only when Configured is true.
	PagesDirectories []string
}

// decodeNoHtmlLinkForPagesOptions reads the option element and anchors each relative directory to
// the config file's directory.
//
// # Anchored to the config, not to the working directory
//
// Upstream hands the string straight to `fs.existsSync`, so a relative path resolves against the
// process's working directory, which is wherever ESLint happened to be started. A committed config is
// read from every directory a person or a CI job starts in, and a path that means something
// different in each is the inert-rule defect with a path in front of it. So a relative directory here
// resolves against the directory of the config file it was written in, the way a tsconfig's paths
// do and the way every other cohere rule taking a path anchors it. When the config sits at the
// project root and ESLint is started there, which is how every consumer runs it, the two readings are
// the same directory. `TestNoHtmlLinkForPagesAnchorsRelativeDirectoriesToTheConfig` pins the choice.
//
// An element written as "" is kept as "", not anchored. Anchored, it would name the config's own
// directory and walk the whole project as a pages directory; upstream finds no directory named ""
// and reads nothing from it, and so does the rule here.
func decodeNoHtmlLinkForPagesOptions(raw []byte, base rule.OptionsBase) (any, error) {
	if len(raw) == 0 {
		return NoHtmlLinkForPagesOptions{}, nil
	}

	// Go whitespace: raw JSON bytes of a rule's options, whose whitespace is the same in both sets.
	trimmed := bytes.TrimSpace(raw)
	var written []string
	switch {
	case len(trimmed) > 0 && trimmed[0] == '"':
		var directory string
		if err := json.Unmarshal(trimmed, &directory); err != nil {
			return nil, fmt.Errorf("expected a pages directory or a list of them: %w", err)
		}
		if directory == "" {
			return NoHtmlLinkForPagesOptions{}, nil
		}
		written = []string{directory}

	case len(trimmed) > 0 && trimmed[0] == '[':
		var elements []json.RawMessage
		if err := json.Unmarshal(trimmed, &elements); err != nil {
			return nil, fmt.Errorf("expected a pages directory or a list of them: %w", err)
		}
		written = make([]string, 0, len(elements))
		seen := make(map[string]bool, len(elements))
		for index, element := range elements {
			var directory string
			// Go whitespace: raw JSON bytes of a rule's options, whose whitespace is the same in both sets.
			if err := json.Unmarshal(element, &directory); err != nil || bytes.TrimSpace(element)[0] != '"' {
				return nil, fmt.Errorf("element %d of the pages directory list is %s, and each must be a string", index+1, element)
			}
			// Upstream's schema says uniqueItems, so ESLint refuses the config before the rule runs.
			if seen[directory] {
				return nil, fmt.Errorf("the pages directory list names %q twice, and its entries must be unique", directory)
			}
			seen[directory] = true
			written = append(written, directory)
		}

	default:
		return nil, fmt.Errorf("expected a pages directory or a list of them, got %s", trimmed)
	}

	anchored := make([]string, 0, len(written))
	for _, directory := range written {
		switch {
		case directory == "" || filepath.IsAbs(directory):
		case base.ConfigDirectory == "":
			return nil, fmt.Errorf("the pages directory %q is relative and the config's directory is not known, "+
				"so there is nothing to resolve it against", directory)
		default:
			directory = filepath.Join(base.ConfigDirectory, directory)
		}
		anchored = append(anchored, directory)
	}
	return NoHtmlLinkForPagesOptions{Configured: true, PagesDirectories: anchored}, nil
}
