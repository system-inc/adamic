// Package spec identifies the bridge's compiler-generated library functions.
package spec

import "strings"

const Prefix = "@tsgo:"

func Name(function string) (string, bool) {
	name, found := strings.CutPrefix(function, Prefix)
	if !found {
		return "", false
	}
	switch name {
	case "tsgoProgram", "tsgoQuery", "tsgoRelease", "tsgoTypeParts", "tsgoInspect":
		return name, true
	}
	return "", false
}
