package registry

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/system-inc/cohere/internal/lint/optionschema"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// CheckSettings refuses a project's `settings` that a run would not read exactly as written, before
// any rule decodes its options.
//
// Settings are read for the plugins whose rule packages register them (rule.RegisterSettings), today
// better-tailwindcss alone, and the rest of this file holds them to what a rule's own options are held
// to (#gj5nm6e):
//
//   - A namespace no package reads is refused by name. `settings.react` would load clean and change
//     nothing, since cohere carries no React settings.
//   - Two spellings of one plugin's namespace are refused together. Upstream reads the first it finds
//     and ignores the other whole.
//   - A key no rule of the plugin reads is refused by name, an upstream option cohere has not ported
//     included: until its port lands, accepting it would be accepting and ignoring it.
//   - What each rule reads is decoded strictly and checked against that rule's upstream schema, the
//     checks its own option element gets, so a value the rule would refuse in its options is refused
//     in settings too, naming the rule.
//
// Upstream itself validates nothing under settings. Refusing here is cohere's direction everywhere: a
// value that would be dropped stops the run with its name instead.
func CheckSettings(base rule.OptionsBase) error {
	if len(base.Settings) == 0 {
		return nil
	}
	readers := map[string]rule.SettingsRegistration{}
	for _, registration := range rule.RegisteredSettings() {
		for _, namespace := range registration.Namespaces {
			readers[namespace] = registration
		}
	}

	namespaces := make([]string, 0, len(base.Settings))
	for namespace := range base.Settings {
		namespaces = append(namespaces, namespace)
	}
	sort.Strings(namespaces)

	var unread []string
	for _, namespace := range namespaces {
		if _, isRead := readers[namespace]; !isRead {
			unread = append(unread, strconv.Quote(namespace))
		}
	}
	if len(unread) > 0 {
		return fmt.Errorf("settings names %s, which no rule cohere runs reads, so the value would be ignored "+
			"silently: cohere reads settings for %s only. Remove it",
			strings.Join(unread, ", "), strings.Join(readNamespaces(readers), ", "))
	}

	for _, registration := range rule.RegisteredSettings() {
		var written []string
		for _, namespace := range registration.Namespaces {
			if _, isWritten := base.Settings[namespace]; isWritten {
				written = append(written, namespace)
			}
		}
		if len(written) > 1 {
			return fmt.Errorf("settings names both %q and %q, two spellings of one plugin's settings: "+
				"upstream reads %q and ignores the other whole, so keep one", written[0], written[1], written[0])
		}
		if len(written) == 0 {
			continue
		}
		if err := checkNamespace(registration, written[0], base); err != nil {
			return err
		}
	}
	return nil
}

// checkNamespace decodes and schema-checks the part of one namespace's settings each rule reads.
func checkNamespace(registration rule.SettingsRegistration, namespace string, base rule.OptionsBase) error {
	parts, err := registration.Split(base.Settings[namespace])
	if err != nil {
		return fmt.Errorf("settings[%q]: %w", namespace, err)
	}
	decoders := map[string]func(raw []byte, base rule.OptionsBase) (any, error){}
	for _, registered := range rule.Registered() {
		decoders[registered.Rule.Name] = registered.DecodeAt
	}
	// Decoded against the anchors alone, so the decoder reads the part as it would an option element
	// rather than merging the settings into it a second time.
	anchors := rule.OptionsBase{ConfigDirectory: base.ConfigDirectory, ProjectRoot: base.ProjectRoot}

	ruleNames := make([]string, 0, len(parts))
	for ruleName := range parts {
		ruleNames = append(ruleNames, ruleName)
	}
	sort.Strings(ruleNames)
	for _, ruleName := range ruleNames {
		part := parts[ruleName]
		if decode := decoders[ruleName]; decode != nil {
			if _, err := decode(part, anchors); err != nil {
				return fmt.Errorf("settings[%q], as %s reads it: %w", namespace, ruleName, err)
			}
		}
		if err := optionschema.Check(ruleName, []json.RawMessage{part}); err != nil {
			return fmt.Errorf("settings[%q], as %s reads it: %w", namespace, ruleName, err)
		}
	}
	return nil
}

// readNamespaces is every namespace some package reads, sorted and quoted, for the refusal's message.
func readNamespaces(readers map[string]rule.SettingsRegistration) []string {
	names := make([]string, 0, len(readers))
	for namespace := range readers {
		names = append(names, strconv.Quote(namespace))
	}
	sort.Strings(names)
	return names
}
