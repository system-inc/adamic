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
	sharedClosurePrototypes := map[string]bool{}
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
		if !d.function && d.initializer < 0 && len(tokens) == 3 {
			switch tokens[0].text {
			case "adamic_code_function", "adamic_counted_code_function", "adamic_method_function", "adamic_counted_method_function":
				count := ""
				if strings.Contains(tokens[0].text, "counted") {
					count = ", size_t argument_count"
				}
				receiver := "adamic_closure"
				if strings.Contains(tokens[0].text, "method") {
					receiver = "adamic_object"
				}
				declaration = fmt.Sprintf("adamic_value %s(%s *self, adamic_value *arguments%s);\n", names[d.name], receiver, count)
				sharedClosurePrototypes[d.name] = true
			}
		}

		if old, ok := definitions[d.name]; ok {
			if old.declaration != declaration {
				return "", nil, fmt.Errorf("native: split: conflicting declaration %s: %q versus %q", d.name, old.declaration, declaration)
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
			// Keep global storage beside its initialization and cleanup in main.
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
	for _, name := range ordered {
		if sharedClosurePrototypes[name] {
			common = strings.TrimSuffix(common, "#endif\n") + definitions[name].declaration + "#endif\n"
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
				if !sharedClosurePrototypes[name] {
					header.WriteString(definitions[name].declaration)
				}
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

// A relocation names the full declaration ABI. A caller with a different canonical
// declaration cannot resolve the defining unit's guard, even without LTO or sanitizers.
func declarationGuard(name, declaration string) string {
	digest := sha256.Sum256([]byte(name + "\x00" + declaration))
	return fmt.Sprintf("adamic_declaration_%x", digest[:])
}

// Async frames and scheduler helpers belong to the source function, never its
// position in the IR. Only identifier tokens are rewritten.
func (e *emitter) canonicalAsyncDeclarations(source string, index int) string {
	tokens, err := cTokens(source)
	if err != nil {
		panic(err)
	}
	replacements := map[string]string{}
	for _, prefix := range []string{"adamic_generated_frame_", "adamic_async_children_", "adamic_async_finish_", "adamic_async_abandon_", "adamic_async_resume_"} {
		replacements[fmt.Sprintf("%s%d", prefix, index)] = prefix + e.functionName(index)
	}
	var out strings.Builder
	previous := 0
	for _, token := range tokens {
		if replacement, ok := replacements[token.text]; ok {
			out.WriteString(source[previous:token.start])
			out.WriteString(replacement)
			previous = token.end
		}
	}
	out.WriteString(source[previous:])
	source = out.String()
	declarations, err := splitDeclarations(source)
	if err != nil {
		panic(err)
	}
	out.Reset()
	previous = 0
	for _, declaration := range declarations {
		if declaration.function {
			at := declaration.tokens[0].start
			out.WriteString(source[previous:at])
			fmt.Fprintf(&out, "// adamic-module %q\n", e.program.Functions[index].Source.Module)
			previous = at
		}
	}
	out.WriteString(source[previous:])
	return out.String()
}
