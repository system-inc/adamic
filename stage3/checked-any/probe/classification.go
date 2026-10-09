package main

import (
	"fmt"
	"regexp"
	"strconv"
)

var diagnosticPosition = regexp.MustCompile(`^(.+):(\d+):(\d+):`)

func insideDeclaration(reason, source string, first, last int) bool {
	position := diagnosticPosition.FindStringSubmatch(reason)
	if len(position) != 4 || position[1] != source {
		return false
	}
	line, err := strconv.Atoi(position[2])
	return err == nil && line >= first && line <= last
}

// A lowered type or transport does not establish the check at a later consumer.
// Keep its outside-use obligation explicit instead of promoting it to checked.
func unobservedUse(record map[string]any) (string, string) {
	name, _ := record["declaration_name"].(string)
	kind, _ := record["declaration_kind"].(string)
	binding, _ := record["binding"].(string)
	location := fmt.Sprintf("%s:%v:%v", record["file"], record["line"], record["column"])
	if kind == "InterfaceDeclaration" || kind == "TypeAliasDeclaration" {
		return fmt.Sprintf("%s: type-only declaration %s has no executable use; its concrete consumers are outside the extracted declaration", location, name), "type_only_consumer_outside"
	}
	if generic, _ := record["generic_declaration"].(bool); generic {
		return fmt.Sprintf("%s: generic declaration %s is registered but has no concrete caller type arguments in the extracted program; its body requires an instantiation outside the declaration", location, name), "generic_instantiation_outside"
	}
	refs, _ := record["references"].([]any)
	for _, r := range refs {
		ref := r.(map[string]any)
		if inside, _ := ref["within_declaration"].(bool); !inside {
			return fmt.Sprintf("%s: %s has no attributed guard in declaration %s; its first outside reference is %s:%v:%v (%s); probing that consumer requires its enclosing declaration", location, binding, name, record["file"], ref["line"], ref["column"], ref["kind"]), "consumer_outside_declaration"
		}
	}
	return fmt.Sprintf("%s: declaration %s lowered without a guard attributed to %s; no outside reference is present in this stock file, so a caller or dataflow consumer must be supplied before this site can be classified checked", location, name, binding), "consumer_not_in_stock_file"
}
