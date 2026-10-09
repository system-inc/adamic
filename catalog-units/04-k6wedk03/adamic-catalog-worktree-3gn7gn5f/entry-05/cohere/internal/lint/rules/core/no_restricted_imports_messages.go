package core

import (
	"fmt"
	"strings"

	"github.com/system-inc/cohere/internal/lint/rule"
)

// Upstream carries twenty-two message ids for this rule, and every one of them is a PAIR: the same
// sentence with and without the project's own `message` appended. The suffix is what varies, never
// the reasoning, so each builder below takes the custom message and picks between the two ids.
//
// The ids are kept exactly as upstream spells them, because a suppression comment or a metrics
// dashboard keyed on `pathWithCustomMessage` has to keep resolving across the port.

// noRestrictedImportsWithCustomMessage appends a project's own message and picks the paired id.
//
// The two ids are not interchangeable even though they render nearly the same text: upstream's
// `pathWithCustomMessage` exists so a reader can tell a project's guidance from ESLint's, and a port
// collapsing them would report a message the config author wrote under an id that says it did not.
func noRestrictedImportsWithCustomMessage(
	plainId string,
	customId string,
	description string,
	customMessage string,
) rule.Message {
	if customMessage == "" {
		return rule.Message{Id: plainId, Description: description}
	}
	return rule.Message{Id: customId, Description: description + " " + customMessage}
}

// noRestrictedImportsFormatNames renders a name list the way upstream's `Intl.ListFormat("en-US")`
// does: quoted, comma separated, with `and` before the last.
//
// Reproduced rather than simplified because the rendered text is what a reader sees and what the
// fixtures assert. `Intl.ListFormat` puts no serial comma before `and` for two items and does for
// three or more, which is the shape below.
func noRestrictedImportsFormatNames(names []string) string {
	quoted := make([]string, 0, len(names))
	for _, name := range names {
		quoted = append(quoted, "'"+name+"'")
	}
	switch len(quoted) {
	case 0:
		return ""
	case 1:
		return quoted[0]
	case 2:
		return quoted[0] + " and " + quoted[1]
	default:
		return strings.Join(quoted[:len(quoted)-1], ", ") + ", and " + quoted[len(quoted)-1]
	}
}

// noRestrictedImportsIsOrAre agrees the verb with the list's length, as upstream's `isOrAre` does.
func noRestrictedImportsIsOrAre(names []string) string {
	if len(names) == 1 {
		return "is"
	}
	return "are"
}

// noRestrictedImportsPathMessage is upstream's `path` / `pathWithCustomMessage`.
func noRestrictedImportsPathMessage(source string, customMessage string) rule.Message {
	return noRestrictedImportsWithCustomMessage("path", "pathWithCustomMessage",
		fmt.Sprintf("`%s` is restricted from being imported here. The project has banned this "+
			"module at its dependency boundary, usually because something else already wraps it "+
			"or because depending on it from this layer inverts the intended direction. Import "+
			"whatever the project provides in its place.", source),
		customMessage)
}

// noRestrictedImportsPatternMessage is upstream's `patterns` / `patternWithCustomMessage`.
func noRestrictedImportsPatternMessage(source string, customMessage string) rule.Message {
	return noRestrictedImportsWithCustomMessage("patterns", "patternWithCustomMessage",
		fmt.Sprintf("`%s` is restricted from being imported here by a pattern. A pattern bans a "+
			"whole shape of module rather than one name, so the fix is usually to import from the "+
			"module's public entry point rather than reaching into it.", source),
		customMessage)
}

// noRestrictedImportsImportNameMessage is upstream's `importName` / `importNameWithCustomMessage`.
func noRestrictedImportsImportNameMessage(
	source string,
	importName string,
	customMessage string,
) rule.Message {
	return noRestrictedImportsWithCustomMessage("importName", "importNameWithCustomMessage",
		fmt.Sprintf("`%s` imported from `%s` is restricted. The module itself is fine to import; "+
			"this one binding is not. Take what you need from the rest of the module, or from "+
			"whatever the project offers in its place.", importName, source),
		customMessage)
}

// noRestrictedImportsPatternAndImportNameMessage is upstream's `patternAndImportName` pair.
func noRestrictedImportsPatternAndImportNameMessage(
	source string,
	importName string,
	customMessage string,
) rule.Message {
	return noRestrictedImportsWithCustomMessage(
		"patternAndImportName", "patternAndImportNameWithCustomMessage",
		fmt.Sprintf("`%s` imported from `%s` is restricted by a pattern. The pattern names a "+
			"binding rather than a module, so every module of this shape is affected and renaming "+
			"the import will not help.", importName, source),
		customMessage)
}

// noRestrictedImportsAllowedImportNameMessage is upstream's `allowedImportName` pair.
//
// This is the inverse-list message: the module has an allow-list and this name is not on it.
func noRestrictedImportsAllowedImportNameMessage(
	source string,
	importName string,
	allowed []string,
	customMessage string,
) rule.Message {
	return noRestrictedImportsWithCustomMessage(
		"allowedImportName", "allowedImportNameWithCustomMessage",
		fmt.Sprintf("`%s` imported from `%s` is restricted because only %s %s allowed. The module "+
			"has an allow-list rather than a ban-list, so anything not named on it is closed by "+
			"default and adding a binding is a deliberate decision.",
			importName, source, noRestrictedImportsFormatNames(allowed),
			noRestrictedImportsIsOrAre(allowed)),
		customMessage)
}

// noRestrictedImportsAllowedNamePatternMessage is upstream's `allowedImportNamePattern` pair.
func noRestrictedImportsAllowedNamePatternMessage(
	source string,
	importName string,
	allowedPattern string,
	customMessage string,
) rule.Message {
	return noRestrictedImportsWithCustomMessage(
		"allowedImportNamePattern", "allowedImportNamePatternWithCustomMessage",
		fmt.Sprintf("`%s` imported from `%s` is restricted because only imports that match the "+
			"pattern `%s` are allowed from `%s`.", importName, source, allowedPattern, source),
		customMessage)
}

// noRestrictedImportsEverythingMessage is upstream's `everything` pair, for a namespace import
// against a restrict-list.
func noRestrictedImportsEverythingMessage(
	source string,
	restricted []string,
	customMessage string,
) rule.Message {
	return noRestrictedImportsWithCustomMessage("everything", "everythingWithCustomMessage",
		fmt.Sprintf("A `*` import is invalid because %s from `%s` %s restricted. A namespace "+
			"import binds every export at once, so it cannot avoid the restricted ones. Import the "+
			"bindings you actually use by name.",
			noRestrictedImportsFormatNames(restricted), source,
			noRestrictedImportsIsOrAre(restricted)),
		customMessage)
}

// noRestrictedImportsPatternEverythingMessage is upstream's `patternAndEverything` pair.
func noRestrictedImportsPatternEverythingMessage(
	source string,
	restricted []string,
	customMessage string,
) rule.Message {
	return noRestrictedImportsWithCustomMessage(
		"patternAndEverything", "patternAndEverythingWithCustomMessage",
		fmt.Sprintf("A `*` import is invalid because %s from `%s` %s restricted by a pattern. A "+
			"namespace import binds every export at once, so it cannot avoid the restricted ones.",
			noRestrictedImportsFormatNames(restricted), source,
			noRestrictedImportsIsOrAre(restricted)),
		customMessage)
}

// noRestrictedImportsPatternEverythingWithRegexMessage is upstream's
// `patternAndEverythingWithRegexImportName` pair, where the restriction is a pattern over names
// rather than a list of them.
func noRestrictedImportsPatternEverythingWithRegexMessage(
	source string,
	importNamePattern string,
	customMessage string,
) rule.Message {
	return noRestrictedImportsWithCustomMessage(
		"patternAndEverythingWithRegexImportName",
		"patternAndEverythingWithRegexImportNameAndCustomMessage",
		fmt.Sprintf("A `*` import is invalid because import names matching the pattern `%s` from "+
			"`%s` are restricted. A namespace import binds every export at once, so it cannot "+
			"avoid the ones the pattern names.", importNamePattern, source),
		customMessage)
}

// noRestrictedImportsEverythingWithAllowMessage is upstream's `everythingWithAllowImportNames` pair.
func noRestrictedImportsEverythingWithAllowMessage(
	source string,
	allowed []string,
	customMessage string,
) rule.Message {
	return noRestrictedImportsWithCustomMessage(
		"everythingWithAllowImportNames", "everythingWithAllowImportNamesAndCustomMessage",
		fmt.Sprintf("A `*` import is invalid because only %s from `%s` %s allowed. A namespace "+
			"import binds every export, which is strictly more than the allow-list permits.",
			noRestrictedImportsFormatNames(allowed), source,
			noRestrictedImportsIsOrAre(allowed)),
		customMessage)
}

// noRestrictedImportsEverythingWithAllowedPatternMessage is upstream's
// `everythingWithAllowedImportNamePattern` pair.
func noRestrictedImportsEverythingWithAllowedPatternMessage(
	source string,
	allowedPattern string,
	customMessage string,
) rule.Message {
	return noRestrictedImportsWithCustomMessage(
		"everythingWithAllowedImportNamePattern",
		"everythingWithAllowedImportNamePatternWithCustomMessage",
		fmt.Sprintf("A `*` import is invalid because only imports matching the pattern `%s` from "+
			"`%s` are allowed. A namespace import binds every export, which is strictly more than "+
			"the pattern permits.", allowedPattern, source),
		customMessage)
}
