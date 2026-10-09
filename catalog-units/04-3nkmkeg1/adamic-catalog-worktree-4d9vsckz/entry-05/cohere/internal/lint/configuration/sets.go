package configuration

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"
)

// The house rule sets, carried in the binary so a repository that installs cohere gets them with no
// checkout of our libraries (#njhfftt). A configuration names one in `extends` as `cohere:<name>`, and
// the name is the file's path under sets/ without `.json`: `cohere:typescript`,
// `cohere:system-inc/structure`.
//
//go:embed sets
var embeddedSetFiles embed.FS

// setFiles is what the sets are read from: the embedded tree, which a test replaces with its own.
var setFiles fs.ReadFileFS = embeddedSetFiles

// SetPrefix marks an `extends` entry as a rule set cohere carries rather than a path. Never a path:
// no file a project could write is named with it, so the two readings cannot collide.
const SetPrefix = "cohere:"

// IsSet reports whether a source names an embedded rule set rather than a file on disk.
func IsSet(source string) bool {
	return strings.HasPrefix(source, SetPrefix)
}

// OurTierPrefix names the sets that make a configuration one of ours: `cohere:system-inc/structure`,
// `cohere:system-inc/base`.
const OurTierPrefix = SetPrefix + "system-inc/"

// InOurTiers reports whether a configuration's chain (its sources, as SourcesOf returns them) reaches
// one of our tier sets.
//
// One predicate for the two readers that change behavior on it (#bfxz13m). Inside our tiers the lint
// loader refuses an off or a departure with no reason, and the format resolver takes the format block
// from the Nexus tier alone. Outside them a project goes its own way: an off needs no reason and its
// own `format` block applies. Two predicates would let a project be strict for one reader and free for
// the other.
func InOurTiers(sources []string) bool {
	for _, source := range sources {
		if strings.HasPrefix(source, OurTierPrefix) {
			return true
		}
	}
	return false
}

// InOurTiers reports whether this configuration's chain reaches one of our tier sets. See InOurTiers.
func (c *Config) InOurTiers() bool {
	return c != nil && InOurTiers(c.Sources)
}

// SourceContents returns the bytes of one source in a configuration's chain: an embedded set's text,
// or a file's.
//
// Everything that keys on a configuration's bytes reads them through this, so a chain ending in a set
// hashes the set's text exactly as it would a base file's.
func SourceContents(source string) ([]byte, error) {
	if !IsSet(source) {
		return os.ReadFile(source)
	}
	contents, err := setFiles.ReadFile(setFilePath(strings.TrimPrefix(source, SetPrefix)))
	if err != nil {
		return nil, fmt.Errorf("%s is not a rule set cohere carries; the sets are %s",
			source, strings.Join(SetNames(), ", "))
	}
	return contents, nil
}

// SourcesOnDisk returns the sources that are files, leaving out the embedded sets.
//
// For a caller that stats or watches what it was handed: a set has no file, and it changes only when
// cohere does, which every cache already keys on.
func SourcesOnDisk(sources []string) []string {
	onDisk := make([]string, 0, len(sources))
	for _, source := range sources {
		if !IsSet(source) {
			onDisk = append(onDisk, source)
		}
	}
	return onDisk
}

// SetNames returns every embedded set's name, prefixed and sorted.
func SetNames() []string {
	var names []string
	// The walk reads only the embedded tree, which the build checked exists, so it cannot fail.
	_ = fs.WalkDir(setFiles, "sets", func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".json") {
			return err
		}
		names = append(names, SetPrefix+strings.TrimSuffix(strings.TrimPrefix(path, "sets/"), ".json"))
		return nil
	})
	sort.Strings(names)
	return names
}

func setFilePath(name string) string {
	return "sets/" + name + ".json"
}

// extendsList is what `extends` names: one source, or a list of them, applied in order.
//
// A list is how a project composes sets that each sit on the same base (`cohere:react` and
// `cohere:next` both extend `cohere:typescript`). The shared base is read once, where it first
// appears, so it stays the outermost layer.
type extendsList []string

func (list *extendsList) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		if single != "" {
			*list = extendsList{single}
		}
		return nil
	}
	var many []string
	if err := json.Unmarshal(data, &many); err != nil {
		return fmt.Errorf("\"extends\" names a path or a rule set, or a list of them: %w", err)
	}
	for _, entry := range many {
		if entry == "" {
			return fmt.Errorf("\"extends\" lists an empty entry")
		}
	}
	*list = many
	return nil
}
