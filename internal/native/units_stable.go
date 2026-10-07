package native

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type unitDefinition struct {
	name, declaration, definition, owner string
	references                           []string
	function                             bool
	source                               bool
}

// Full paths determine identity. The escaped spelling is only a readable hint.
func moduleUnit(module string) string {
	digest := sha256.Sum256([]byte(module))
	var escaped strings.Builder
	for _, c := range module {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' {
			escaped.WriteRune(c)
		} else {
			fmt.Fprintf(&escaped, "_%x_", c)
		}
	}
	hint := escaped.String()
	if len(hint) > 80 {
		hint = hint[:80]
	}
	return fmt.Sprintf("module_%s_%x.c", hint, digest[:])
}

func moduleMarker(text string) (string, bool, error) {
	at := strings.LastIndex(text, "// adamic-module ")
	if at < 0 {
		return "", false, nil
	}
	value := strings.SplitN(text[at+len("// adamic-module "):], "\n", 2)[0]
	module, err := strconv.Unquote(strings.TrimSpace(value))
	if err != nil {
		return "", true, fmt.Errorf("native: split: invalid module marker: %w", err)
	}
	return module, true, nil
}

func stableSplitC(source string) (string, []compilationUnit, error) {
	declarations, err := splitDeclarations(source)
	if err != nil {
		return "", nil, err
	}
	names := map[string]string{}
	for _, d := range declarations {
		if d.name != "" && d.name != "main" {
			names[d.name] = "adamic_unit_" + d.name
		}
	}
	rewrite := func(tokens []cToken) string {
		if len(tokens) == 0 {
			return ""
		}
		var out strings.Builder
		position := tokens[0].start
		for _, token := range tokens {
			out.WriteString(source[position:token.start])
			if name, ok := names[token.text]; ok {
				out.WriteString(name)
			} else {
				out.WriteString(token.text)
			}
			position = token.end
		}
		return out.String()
	}
	definitions := map[string]*unitDefinition{}
	common := "#ifndef ADAMIC_UNITS_H\n#define ADAMIC_UNITS_H\n"
	previous := 0
	for _, d := range declarations {
		begin := d.tokens[0].start
		module, marked, err := moduleMarker(source[previous:begin])
		if err != nil {
			return "", nil, err
		}
		previous = d.tokens[len(d.tokens)-1].end
		if d.name == "" {
			common += rewrite(d.tokens) + "\n"
			continue
		}
		tokens := d.tokens
		if tokens[0].text == "static" {
			tokens = tokens[1:]
		}
		declaration := ""
		if d.function {
			signature := d.tokens[:d.body]
			if signature[0].text == "static" {
				signature = signature[1:]
			}
			declaration = rewrite(signature) + ";\n"
		} else {
			limit := len(d.tokens) - 1
			if d.initializer >= 0 {
				limit = d.initializer
			}
			declaration = "extern " + rewrite(d.tokens[1:limit]) + ";\n"
			if d.initializer < 0 && strings.Contains(declaration, "(") {
				declaration = rewrite(d.tokens[1:limit]) + ";\n"
			}
		}
		if old, ok := definitions[d.name]; ok {
			if old.declaration != declaration {
				return "", nil, fmt.Errorf("native: split: conflicting declaration %s", d.name)
			}
			if !d.function && d.initializer < 0 && strings.HasPrefix(d.name, "adamic_class_") {
				continue
			}
			if old.definition != "" {
				return "", nil, fmt.Errorf("native: split: duplicate definition %s", d.name)
			}
		} else {
			definitions[d.name] = &unitDefinition{name: d.name, declaration: declaration}
		}
		item := definitions[d.name]
		// Function prototypes and class forward declarations are declarations, not definitions.
		if !d.function && d.initializer < 0 && (strings.Contains(declaration, "(") || strings.HasPrefix(d.name, "adamic_class_")) {
			continue
		}
		item.definition = rewrite(tokens) + "\n"
		item.function = d.function
		item.source = d.function && marked && module != ""
		if d.name == "main" {
			item.owner = "main.c"
		} else if d.function && marked {
			if module != "" {
				item.owner = moduleUnit(module)
			} else {
				digest := sha256.Sum256([]byte(d.name))
				item.owner = fmt.Sprintf("helpers_%02d.c", int(digest[0])%16)
			}
		} else if !d.function && strings.HasPrefix(d.name, "adamic_global_") {
			// Unmarked forwarders are initialized by the coordinator. Module-ready writes
			// below replace this fallback with the source module's ownership.
			item.owner = "main.c"
		} else if !d.function && (strings.HasPrefix(d.name, "adamic_string_") || strings.HasPrefix(d.name, "adamic_shape_") || strings.HasPrefix(d.name, "adamic_regex_") || strings.Contains(" "+declaration, " const ")) {
			// Content-addressed descriptors may gain consumers without migrating their identity.
			digest := sha256.Sum256([]byte(d.name))
			item.owner = fmt.Sprintf("shared_%02d.c", int(digest[0])%32)
		}
		references := map[string]bool{}
		for _, token := range d.tokens {
			if _, ok := names[token.text]; ok && token.text != d.name {
				references[token.text] = true
			}
		}
		for name := range references {
			item.references = append(item.references, name)
		}
		sort.Strings(item.references)
	}
	// Declaration-owned adapters follow the first source declarations reached through
	// adapter calls. Stop at a source function rather than traversing its implementation.
	for _, item := range definitions {
		if !item.function || item.owner != "" {
			continue
		}
		owners := map[string]bool{}
		visited := map[string]bool{}
		pending := append([]string(nil), item.references...)
		for len(pending) > 0 {
			reference := pending[len(pending)-1]
			pending = pending[:len(pending)-1]
			if visited[reference] {
				continue
			}
			visited[reference] = true
			target, ok := definitions[reference]
			if !ok || !target.function {
				continue
			}
			if target.source {
				owners[target.owner] = true
			} else {
				pending = append(pending, target.references...)
			}
		}
		if len(owners) == 1 {
			for owner := range owners {
				item.owner = owner
			}
		} else {
			digest := sha256.Sum256([]byte(item.name))
			item.owner = fmt.Sprintf("helpers_%02d.c", int(digest[0])%16)
		}
	}
	if main, ok := definitions["main"]; ok {
		coordinator, helpers, err := splitModuleMain(main.definition, names)
		if err != nil {
			return "", nil, err
		}
		// Ready writes mark a global's declaration, rather than later assignments.
		for _, helper := range helpers {
			tokens, _ := cTokens(helper.definition)
			for index, token := range tokens {
				if !strings.HasPrefix(token.text, "adamic_unit_adamic_global_") || !strings.HasSuffix(token.text, "_ready") || index+2 >= len(tokens) || tokens[index+1].text != "=" || tokens[index+2].text != "true" {
					continue
				}
				ready := strings.TrimPrefix(token.text, "adamic_unit_")
				for _, name := range []string{ready, strings.TrimSuffix(ready, "_ready")} {
					if item, ok := definitions[name]; ok {
						item.owner = helper.owner
					}
				}
			}
		}
		main.definition = coordinator
		main.references = nil
		tokens, err := cTokens(coordinator)
		if err != nil {
			return "", nil, err
		}
		for name, rewritten := range names {
			for _, token := range tokens {
				if token.text == rewritten {
					main.references = append(main.references, name)
					break
				}
			}
		}
		for name, item := range helpers {
			definitions[name] = item
			main.references = append(main.references, name)
		}
	}
	common += "#endif\n"
	// Propagate consumers through initializers and adapter bodies. An exclusively used object
	// joins its consumer's module; shared objects retain exactly one separately owned definition.
	consumers := map[string]map[string]bool{}
	for name, item := range definitions {
		consumers[name] = map[string]bool{}
		if item.owner != "" {
			consumers[name][item.owner] = true
		}
	}
	for changed := true; changed; {
		changed = false
		for name, item := range definitions {
			for _, reference := range item.references {
				target, ok := definitions[reference]
				if !ok {
					continue
				}
				if target.owner != "" {
					continue
				}
				for owner := range consumers[name] {
					if !consumers[reference][owner] {
						consumers[reference][owner] = true
						changed = true
					}
				}
			}
		}
	}
	ordered := make([]string, 0, len(definitions))
	for name := range definitions {
		ordered = append(ordered, name)
	}
	sort.Strings(ordered)
	for _, name := range ordered {
		item := definitions[name]
		if item.definition == "" {
			return "", nil, fmt.Errorf("native: split: missing definition %s", name)
		}
		if item.owner == "" {
			if len(consumers[name]) == 1 {
				for owner := range consumers[name] {
					item.owner = owner
				}
			} else {
				digest := sha256.Sum256([]byte(name))
				item.owner = fmt.Sprintf("shared_%02d.c", int(digest[0])%32)
			}
		}
	}
	bodies := map[string]string{}
	needed := map[string]map[string]bool{}
	for _, name := range ordered {
		item := definitions[name]
		bodies[item.owner] += item.definition
		if needed[item.owner] == nil {
			needed[item.owner] = map[string]bool{}
		}
		needed[item.owner][name] = true // The defining unit checks its own ABI too.
		for _, reference := range item.references {
			needed[item.owner][reference] = true
		}
	}
	unitNames := make([]string, 0, len(bodies))
	for name := range bodies {
		unitNames = append(unitNames, name)
	}
	sort.Strings(unitNames)
	var units []compilationUnit
	for _, unitName := range unitNames {
		var header strings.Builder
		header.WriteString("#include \"units.h\"\n")
		for _, name := range ordered {
			if needed[unitName][name] {
				header.WriteString(definitions[name].declaration)
				if name != "main" {
					fmt.Fprintf(&header, "extern const char %s;\n", declarationGuard(name, definitions[name].declaration))
				}
			}
		}
		var guards strings.Builder
		for _, name := range ordered {
			if name == "main" {
				continue
			}
			item := definitions[name]
			guard := declarationGuard(name, item.declaration)
			if item.owner == unitName {
				fmt.Fprintf(&guards, "const char %s = 0;\n", guard)
			}
		}
		digest := sha256.Sum256([]byte(unitName))
		fmt.Fprintf(&guards, "const char *const adamic_abi_references_%x[] = {\n", digest[:16])
		guards.WriteString("0,\n")
		for _, name := range ordered {
			if name != "main" && needed[unitName][name] {
				fmt.Fprintf(&guards, "&%s,\n", declarationGuard(name, definitions[name].declaration))
			}
		}
		guards.WriteString("};\n")
		units = append(units, compilationUnit{unitName, header.String() + bodies[unitName] + guards.String()})
	}
	return common, units, nil
}

// Initialization stays in source order. Only closed scopes with no local crossing a marked
// boundary can be extracted. The coordinator keeps startup, forwarders and global cleanup.
func splitModuleMain(main string, globals map[string]string) (string, map[string]*unitDefinition, error) {
	type boundary struct {
		at     int
		module string
	}
	var boundaries []boundary
	offset := 0
	for _, line := range strings.SplitAfter(main, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "// adamic-module ") {
			module, _, err := moduleMarker(line)
			if err != nil {
				return "", nil, err
			}
			if module != "" {
				boundaries = append(boundaries, boundary{offset, module})
			}
		}
		offset += len(line)
	}
	helpers := map[string]*unitDefinition{}
	if len(boundaries) == 0 {
		return main, helpers, nil
	}
	tail := strings.LastIndex(main, "\treturn 0;")
	if tail < 0 {
		return "", nil, fmt.Errorf("native: split: unsupported main epilogue")
	}
	// Global releases are the final pairs emitted by releaseGlobals, after scope cleanup.
	cleanup := tail
	lines := strings.SplitAfter(main[:tail], "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	for len(lines) >= 2 {
		last := strings.TrimSpace(lines[len(lines)-1])
		before := strings.TrimSpace(lines[len(lines)-2])
		global := strings.TrimSuffix(last, " = NULL;")
		if last == global || !strings.HasPrefix(global, "adamic_unit_adamic_global_") || before != "adamic_release("+global+");" {
			break
		}
		cleanup -= len(lines[len(lines)-1]) + len(lines[len(lines)-2])
		lines = lines[:len(lines)-2]
	}
	var coordinator strings.Builder
	coordinator.WriteString(main[:boundaries[0].at])
	seenLocals := map[string]bool{}
	prefixTokens, _ := cTokens(main[:boundaries[0].at])
	for _, token := range prefixTokens {
		if strings.HasPrefix(token.text, "adamic_local_") || strings.HasPrefix(token.text, "adamic_temporary_") {
			seenLocals[token.text] = true
		}
	}
	originalNames := map[string]string{}
	for name, rewritten := range globals {
		originalNames[rewritten] = name
	}
	for index, boundary := range boundaries {
		end := cleanup
		if index+1 < len(boundaries) {
			end = boundaries[index+1].at
		}
		body := main[boundary.at:end]
		tokens, err := cTokens(body)
		if err != nil {
			return "", nil, err
		}
		depth := 0
		locals := map[string]bool{}
		references := map[string]bool{}
		for _, token := range tokens {
			switch token.text {
			case "return":
				return "", nil, fmt.Errorf("native: split: return crosses module initialization")
			case "{":
				depth++
			case "}":
				depth--
			}
			if depth < 0 {
				return "", nil, fmt.Errorf("native: split: module initialization crosses a scope")
			}
			if strings.HasPrefix(token.text, "adamic_local_") || strings.HasPrefix(token.text, "adamic_temporary_") {
				if seenLocals[token.text] {
					return "", nil, fmt.Errorf("native: split: initialization local %s crosses modules", token.text)
				}
				locals[token.text] = true
			}
			if name, ok := originalNames[token.text]; ok {
				references[name] = true
			}
		}
		if depth != 0 {
			return "", nil, fmt.Errorf("native: split: module initialization crosses a scope")
		}
		for local := range locals {
			seenLocals[local] = true
		}
		identity := boundary.module
		if identity == "" {
			identity = "<forwarders>"
		}
		digest := sha256.Sum256([]byte(identity))
		name := fmt.Sprintf("adamic_unit_initialize_%x", digest[:])
		if _, ok := helpers[name]; ok {
			return "", nil, fmt.Errorf("native: split: repeated initialization module %q", identity)
		}
		item := &unitDefinition{name: name, owner: moduleUnit(identity), declaration: "void " + name + "(void);\n", definition: "void " + name + "(void) {\n" + body + "}\n"}
		for reference := range references {
			item.references = append(item.references, reference)
		}
		sort.Strings(item.references)
		helpers[name] = item
		fmt.Fprintf(&coordinator, "\t%s();\n", name)
	}
	// Cleanup may refer only to globals, never to extracted locals.
	remaining, _ := cTokens(main[cleanup:])
	for _, token := range remaining {
		if seenLocals[token.text] {
			return "", nil, fmt.Errorf("native: split: local %s survives module initialization", token.text)
		}
	}
	coordinator.WriteString(main[cleanup:])
	return coordinator.String(), helpers, nil
}

// A relocation names the full declaration ABI. A caller with a different canonical
// declaration cannot resolve the defining unit's guard, even without LTO or sanitizers.
func declarationGuard(name, declaration string) string {
	digest := sha256.Sum256([]byte(name + "\x00" + declaration))
	return fmt.Sprintf("adamic_declaration_%x", digest[:])
}
