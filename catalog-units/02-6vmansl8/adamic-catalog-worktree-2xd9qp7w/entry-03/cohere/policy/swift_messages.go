package policy

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// The Swift rules' messages, compiled in as one typed function per message (#hg415t1, the Swift half of
// #xd6f0n6). A function's parameters are exactly the values its Swift text uses, so a rule that skips a
// value, passes one the text has no place for, or names a message the catalog lacks does not compile.
//
// The Swift text is resolved here, at generation: its terms in their Swift words and its string phrases
// in place. The engine never reads the catalog. An object phrase becomes an enum whose cases are its
// options, each carrying the values its own text uses. The function takes that enum, so the rule names
// its choice at the call.
//
// What drifts is caught twice: TestSwiftFilesAreCurrent here, and a Swift test that recomputes
// RuleMessages.sourceDigest from policy/messages/ on disk. The digest covers only what reaches Swift
// (swiftMessagesDigest), so a TypeScript-only edit leaves the generated file alone.

// swiftLineLength is the house Swift format's line length (HouseSwiftFormat.swift).
const swiftLineLength = 120

// swiftMessagesSource is RuleMessages.generated.swift for the catalog, carrying digest as what it was
// generated from.
func swiftMessagesSource(catalog *MessageCatalog, digest string) ([]byte, error) {
	type rule struct {
		name     string
		typeName string
		ids      []string
	}
	var rules []rule
	for ruleName, language := range catalog.languages {
		if language != MessageLanguageSwift {
			continue
		}
		_, leaf, _ := strings.Cut(ruleName, "/")
		ids := make([]string, 0, len(catalog.templates[ruleName]))
		for id := range catalog.templates[ruleName] {
			ids = append(ids, id)
		}
		// A twin file whose messages are all TypeScript's names the Swift rule with nothing for it to render.
		if len(ids) == 0 {
			continue
		}
		sort.Strings(ids)
		rules = append(rules, rule{name: ruleName, typeName: pascalCase(leaf), ids: ids})
	}
	sort.Slice(rules, func(left, right int) bool { return rules[left].typeName < rules[right].typeName })

	var source bytes.Buffer
	source.WriteString("/*\n Generated from cohere's policy/messages/ by `go run ./policy/tools/generate`. Do not edit it here:\n")
	source.WriteString(" edit the message files and run that again. While the two differ, a test in each engine fails.\n */\n")
	source.WriteString("enum RuleMessages {\n")
	source.WriteString("    /* A finding's message: the id the catalog files it under, and its text. */\n")
	source.WriteString("    struct Message: Equatable, Sendable {\n        let id: String\n        let text: String\n    }\n\n")
	source.WriteString("    /* SHA-256 of the messages that reach Swift, resolved, which a test recomputes from policy/messages/ on disk. */\n")
	fmt.Fprintf(&source, "    static let sourceDigest = \"%s\"\n\n", digest)
	source.WriteString("    /* Every message here, as `Rule.id`, for the test that fails on one no rule renders. */\n")
	var entries []string
	for _, each := range rules {
		for _, id := range each.ids {
			entries = append(entries, `"`+each.typeName+"."+id+`"`)
		}
	}
	// The house format puts a trailing comma after the last of several elements and none after a lone one.
	switch len(entries) {
	case 0:
		source.WriteString("    static let all: [String] = []\n")
	case 1:
		fmt.Fprintf(&source, "    static let all = [\n        %s\n    ]\n", entries[0])
	default:
		fmt.Fprintf(&source, "    static let all = [\n        %s,\n    ]\n", strings.Join(entries, ",\n        "))
	}
	for _, each := range rules {
		fmt.Fprintf(&source, "\n    /* %s */\n", each.name)
		fmt.Fprintf(&source, "    enum %s {\n", each.typeName)
		for index, id := range each.ids {
			if index > 0 {
				source.WriteString("\n")
			}
			if err := writeSwiftMessage(&source, id, catalog.templates[each.name][id]); err != nil {
				return nil, fmt.Errorf("%s %q: %w", each.name, id, err)
			}
		}
		source.WriteString("    }\n")
	}
	source.WriteString("}\n")
	return source.Bytes(), nil
}

// writeSwiftMessage writes one message as an enum per object phrase it picks from, then its function.
func writeSwiftMessage(source *bytes.Buffer, id string, template messageTemplate) error {
	slots := orderedSlots(template.text)
	var parameters []string
	textValues := map[string]bool{}
	for _, slot := range slots {
		if slot.phrase {
			parameters = append(parameters, slot.name+": "+pascalCase(slot.name))
			continue
		}
		textValues[slot.name] = true
		parameters = append(parameters, slot.name+": String")
	}

	phraseNames := make([]string, 0, len(template.options))
	for phrase := range template.options {
		phraseNames = append(phraseNames, phrase)
	}
	sort.Strings(phraseNames)
	for _, phrase := range phraseNames {
		options := template.options[phrase]
		optionNames := make([]string, 0, len(options))
		for name := range options {
			optionNames = append(optionNames, name)
		}
		sort.Strings(optionNames)
		fmt.Fprintf(source, "        enum %s {\n", pascalCase(phrase))
		for _, name := range optionNames {
			values := valueSlots(options[name])
			for _, value := range values {
				if textValues[value] {
					return fmt.Errorf("the value %q is in the text and in the option %q of %q, so it has no one place to be passed", value, name, phrase)
				}
			}
			if len(values) == 0 {
				fmt.Fprintf(source, "            case %s\n", name)
				continue
			}
			labelled := make([]string, len(values))
			for index, value := range values {
				labelled[index] = value + ": String"
			}
			fmt.Fprintf(source, "            case %s(%s)\n", name, strings.Join(labelled, ", "))
		}
		source.WriteString("        }\n\n")
	}

	// A signature past the house format's 120 columns takes one parameter to a line, as the format breaks it.
	signature := fmt.Sprintf("        static func %s(%s) -> Message {", id, strings.Join(parameters, ", "))
	if len(signature) > swiftLineLength {
		signature = fmt.Sprintf("        static func %s(\n            %s,\n        ) -> Message {", id, strings.Join(parameters, ",\n            "))
	}
	source.WriteString(signature + "\n")
	for _, phrase := range phraseNames {
		options := template.options[phrase]
		optionNames := make([]string, 0, len(options))
		for name := range options {
			optionNames = append(optionNames, name)
		}
		sort.Strings(optionNames)
		// Each option's text sits on the line below its case, where the house format puts a long one and
		// keeps a short one.
		fmt.Fprintf(source, "            let %sText =\n                switch %s {\n", phrase, phrase)
		for _, name := range optionNames {
			values := valueSlots(options[name])
			literal, err := swiftInterpolatedLiteral(options[name])
			if err != nil {
				return fmt.Errorf("the option %q of %q: %w", name, phrase, err)
			}
			if len(values) == 0 {
				fmt.Fprintf(source, "                    case .%s:\n                        %s\n", name, literal)
				continue
			}
			bound := make([]string, len(values))
			for index, value := range values {
				bound[index] = "let " + value
			}
			fmt.Fprintf(source, "                    case .%s(%s):\n                        %s\n", name, strings.Join(bound, ", "), literal)
		}
		source.WriteString("                }\n")
	}
	text := template.text
	for _, phrase := range phraseNames {
		text = strings.ReplaceAll(text, "<<"+phrase+">>", "{{"+phrase+"Text}}")
	}
	literal, err := swiftInterpolatedLiteral(text)
	if err != nil {
		return err
	}
	// A function of one expression returns it, and one that first picks its options returns explicitly.
	// The call is broken one argument to a line, the house format's shape for a call this long, which it
	// keeps for a shorter one because it respects the breaks it finds.
	call := "Message("
	if len(phraseNames) > 0 {
		call = "return Message("
	}
	fmt.Fprintf(source, "            %s\n                id: \"%s\",\n                text:\n                    %s,\n            )\n", call, id, literal)
	source.WriteString("        }\n")
	return nil
}

// messageSlot is a `{{value}}` or `<<phrase>>` in a message's text.
type messageSlot struct {
	name   string
	phrase bool
}

// orderedSlots are the text's values and phrases in the order they first appear, so a function's
// parameters read in the order of the sentence.
func orderedSlots(text string) []messageSlot {
	var slots []messageSlot
	seen := map[string]bool{}
	for index := 0; index < len(text); index++ {
		for _, kind := range []struct {
			open, close string
			phrase      bool
		}{{"{{", "}}", false}, {"<<", ">>", true}} {
			if !strings.HasPrefix(text[index:], kind.open) {
				continue
			}
			length := strings.Index(text[index+len(kind.open):], kind.close)
			if length < 0 {
				continue
			}
			name := text[index+len(kind.open) : index+len(kind.open)+length]
			if !seen[name] {
				seen[name] = true
				slots = append(slots, messageSlot{name: name, phrase: kind.phrase})
			}
			index += len(kind.open) + length + len(kind.close) - 1
			break
		}
	}
	return slots
}

// valueSlots are the `{{value}}` names in text, in order of first appearance.
func valueSlots(text string) []string {
	var names []string
	for _, slot := range orderedSlots(text) {
		if !slot.phrase {
			names = append(names, slot.name)
		}
	}
	return names
}

// swiftInterpolatedLiteral is text as a one-line Swift raw string, each `{{name}}` interpolating the
// Swift name. The delimiter outruns every run of `#` in the text, so nothing in it ends the string or
// starts an escape, and a line break is written as the raw string's own escape.
func swiftInterpolatedLiteral(text string) (string, error) {
	if strings.ContainsRune(text, '\r') {
		return "", fmt.Errorf("it holds a carriage return")
	}
	delimiter := strings.Repeat("#", longestRun([]byte(text), '#')+1)
	escaped := strings.ReplaceAll(text, "\n", `\`+delimiter+"n")
	for _, name := range valueSlots(text) {
		escaped = strings.ReplaceAll(escaped, "{{"+name+"}}", `\`+delimiter+"("+name+")")
	}
	return delimiter + `"` + escaped + `"` + delimiter, nil
}

// pascalCase is a kebab or camel name as a Swift type name: `consistency-no-bare-throw` becomes
// `ConsistencyNoBareThrow`, `exitCount` becomes `ExitCount`.
func pascalCase(name string) string {
	var result strings.Builder
	for _, word := range strings.Split(name, "-") {
		if word != "" {
			result.WriteString(strings.ToUpper(word[:1]) + word[1:])
		}
	}
	return result.String()
}

// swiftMessagesDigest is SHA-256 over what reaches Swift and nothing else, so a TypeScript-only edit
// leaves RuleMessages.generated.swift byte-identical. For each Swift rule in name order, and each of its
// messages in id order: the rule, the id and the resolved text (its Swift terms and string phrases in
// place), then each object phrase the text picks from, in name order, as `<<name>>` and each option's
// name and text in name order. Every field ends with a zero byte. The Swift test resolves the files on
// disk the same way and compares.
func swiftMessagesDigest(catalog *MessageCatalog) string {
	var rules []string
	for ruleName, language := range catalog.languages {
		if language == MessageLanguageSwift {
			rules = append(rules, ruleName)
		}
	}
	sort.Strings(rules)
	hash := sha256.New()
	field := func(text string) {
		hash.Write([]byte(text))
		hash.Write([]byte{0})
	}
	for _, ruleName := range rules {
		ids := make([]string, 0, len(catalog.templates[ruleName]))
		for id := range catalog.templates[ruleName] {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			template := catalog.templates[ruleName][id]
			field(ruleName)
			field(id)
			field(template.text)
			phrases := make([]string, 0, len(template.options))
			for phrase := range template.options {
				phrases = append(phrases, phrase)
			}
			sort.Strings(phrases)
			for _, phrase := range phrases {
				field("<<" + phrase + ">>")
				options := make([]string, 0, len(template.options[phrase]))
				for name := range template.options[phrase] {
					options = append(options, name)
				}
				sort.Strings(options)
				for _, name := range options {
					field(name)
					field(template.options[phrase][name])
				}
			}
		}
	}
	return hex.EncodeToString(hash.Sum(nil))
}
