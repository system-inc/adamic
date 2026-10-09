package typescript

import (
	"encoding/json"

	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoDeprecatedOptions is the rule's one option: an allowlist of things whose deprecation is
// accepted, so a codebase can keep using a deprecated API it has decided to keep using.
//
// The default is an empty allowlist, which is the zero value, so an absent configuration needs no
// special handling. That is worth stating because it is not true of every rule in this family:
// `no-base-to-string` defaults to four builtin type names and a decoder returning the zero value
// there would change behaviour.
type NoDeprecatedOptions struct {
	// Allow are specifier objects: a name plus where it comes from.
	Allow []type_checking.TypeOrValueSpecifier
	// AllowInline are the bare-string entries, matched against a type's own name.
	AllowInline []string
}

// noDeprecatedRawOptions is the wire shape.
//
// cohere's config layer strips ESLint's `[severity, options]` tuple and stores only the second
// element, so what arrives here is the bare object rather than upstream's one-element array.
type noDeprecatedRawOptions struct {
	Allow []noDeprecatedRawSpecifier `json:"allow"`
}

// noDeprecatedRawSpecifier is one allowlist entry, in either of its two wire forms.
type noDeprecatedRawSpecifier struct {
	inline string

	From    string
	Name    []string
	Path    string
	Package string
}

// UnmarshalJSON accepts a bare string or the specifier object, and `name` as a string or an array.
//
// An unrecognized `from` is dropped rather than erroring, matching `only-throw-error` and
// `no-floating-promises`: the specifier then matches nothing, whereas keeping it would silently mean
// `file`, which is the zero value of the enum.
func (s *noDeprecatedRawSpecifier) UnmarshalJSON(raw []byte) error {
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		*s = noDeprecatedRawSpecifier{inline: asString}
		return nil
	}

	var object struct {
		From    string          `json:"from"`
		Name    json.RawMessage `json:"name"`
		Path    string          `json:"path"`
		Package string          `json:"package"`
	}
	if err := rule.UnmarshalOptions(raw, &object); err != nil {
		return err
	}

	*s = noDeprecatedRawSpecifier{From: object.From, Path: object.Path, Package: object.Package}

	var names []string
	if err := json.Unmarshal(object.Name, &names); err == nil {
		s.Name = names
		return nil
	}
	var name string
	if err := json.Unmarshal(object.Name, &name); err != nil {
		return err
	}
	s.Name = []string{name}
	return nil
}

// DecodeNoDeprecatedOptions maps upstream's JSON onto the struct the rule reads.
//
// Hand-written rather than `rule.DecodeOptionsInto` because `allow` is one heterogeneous array whose
// two forms land in two different fields, and because `from` is an enum rather than a string.
func DecodeNoDeprecatedOptions(raw []byte) (any, error) {
	decoded, err := rule.DecodeOptionsInto[noDeprecatedRawOptions]()(raw)
	if err != nil {
		return NoDeprecatedOptions{}, err
	}

	wire, _ := decoded.(noDeprecatedRawOptions)

	options := NoDeprecatedOptions{}
	for _, entry := range wire.Allow {
		if entry.inline != "" {
			options.AllowInline = append(options.AllowInline, entry.inline)
			continue
		}

		specifier := type_checking.TypeOrValueSpecifier{Name: entry.Name, Path: entry.Path, Package: entry.Package}
		switch entry.From {
		case "file":
			specifier.From = type_checking.TypeOrValueSpecifierFromFile
		case "lib":
			specifier.From = type_checking.TypeOrValueSpecifierFromLib
		case "package":
			specifier.From = type_checking.TypeOrValueSpecifierFromPackage
		default:
			continue
		}
		options.Allow = append(options.Allow, specifier)
	}

	return options, nil
}
