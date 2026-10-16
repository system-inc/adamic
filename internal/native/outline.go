package native

import (
	"fmt"
	"strings"
)

// Keep initializer calls out of clang's inliner, including the unsplit build.
// Otherwise a large literal's stores become one huge backend function again.
const initializerAttribute = " __attribute__((noinline))"
const initializerChunkStatements = 128

type initializerPart struct {
	module     string
	statements []string
}
type initializerChunk struct {
	module, name, body string
	parameters         []string
	part               int
}
type initializerLocal struct {
	name, of    string
	start, end  int
	initialized bool
}

// moduleMain emits the original scopes and statements first. Outlining only moves
// their C into helpers; readiness writes, exception paths, and cleanup stay in order.
func (e *emitter) moduleMain() string {
	prefix := len(e.program.Main)
	for _, module := range e.program.MainModules {
		prefix -= module.Statements
	}
	if prefix < 0 {
		panic("native: invalid main module boundaries")
	}
	starts := map[int]string{0: ""}
	offset := prefix
	for _, module := range e.program.MainModules {
		if module.Statements > 0 {
			starts[offset] = module.Module
		}
		offset += module.Statements
	}
	e.scopes = append(e.scopes, nil)
	var parts []initializerPart
	for index := range e.program.Main {
		if module, ok := starts[index]; ok {
			e.resetCounters()
			readable := module
			if len(readable) > 40 {
				readable = readable[:40]
			}
			e.mainModule = stableName("module", readable, module)
			parts = append(parts, initializerPart{module: module})
		}
		e.statementAt(&e.program.Main[index])
		text := e.out.String()
		e.out.Reset()
		parts[len(parts)-1].statements = append(parts[len(parts)-1].statements, text)
	}
	e.releaseScopes(len(e.scopes) - 1)
	cleanup := e.out.String()
	e.out.Reset()
	e.scopes = e.scopes[:len(e.scopes)-1]
	e.mainModule = ""
	return e.outlineInitializers(parts, cleanup)
}

// Flat generated statements may be cut even within a source initializer (large
// tables are single array literals). Structured control flow stays with its source
// statement so labels, handlers, and lexical lifetimes cannot cross helper frames.
func initializerStatements(source string) []string {
	tokens, err := cTokens(source)
	if err != nil {
		panic(err)
	}
	if len(tokens) > 0 && tokens[0].text == "{" {
		return []string{source}
	}
	for _, token := range tokens {
		switch token.text {
		case "if", "for", "while", "do", "goto", "return":
			return []string{source}
		}
	}
	var statements []string
	depth, paren, begin := 0, 0, 0
	for _, token := range tokens {
		switch token.text {
		case "{":
			depth++
		case "}":
			depth--
		case "(":
			paren++
		case ")":
			paren--
		}
		if token.text == ";" && depth == 0 && paren == 0 {
			statements = append(statements, source[begin:token.end])
			begin = token.end
		}
	}
	if strings.TrimSpace(source[begin:]) != "" {
		panic("native: incomplete initializer statement")
	}
	return statements
}

// Root declarations are the only values that can outlive a generated chunk.
// Nested declarations remain inside the structured statement that owns them.
func initializerLocals(source string) []initializerLocal {
	tokens, err := cTokens(source)
	if err != nil {
		panic(err)
	}
	var locals []initializerLocal
	depth, paren, begin := 0, 0, 0
	for i, token := range tokens {
		if depth == 0 && paren == 0 && i == begin {
			j := i
			for j < len(tokens) && (tokens[j].text == "*" || initializerType(tokens[j].text)) {
				j++
			}
			if j > i && j < len(tokens) && (strings.HasPrefix(tokens[j].text, "adamic_temporary_") || strings.HasPrefix(tokens[j].text, "adamic_local_")) {
				local := initializerLocal{name: tokens[j].text, of: strings.TrimSpace(source[token.start:tokens[j].start]), start: token.start}
				if j+1 >= len(tokens) {
					panic("native: incomplete initializer declaration")
				}
				switch tokens[j+1].text {
				case "=":
					local.initialized = true
					local.end = tokens[j+1].end
				case ";":
					local.end = tokens[j+1].end
				case ",":
					// Runtime iterator key/value pairs are used only within their
					// structured source statement, which is never cut into chunks.
					continue
				default:
					panic("native: unsupported initializer declarator " + local.name)
				}
				locals = append(locals, local)
			}
		}
		switch token.text {
		case "{":
			depth++
		case "}":
			depth--
		case "(":
			paren++
		case ")":
			paren--
		}
		if (token.text == ";" || token.text == "}") && depth == 0 && paren == 0 {
			begin = i + 1
		}
	}
	return locals
}
func initializerType(token string) bool {
	switch token {
	case "double", "bool", "void", "size_t", "int64_t", "adamic_value", "adamic_heap", "adamic_string", "adamic_object", "adamic_array", "adamic_map", "adamic_closure", "adamic_cell", "adamic_weak", "adamic_region", "adamic_map_iterator", "adamic_maybe_number", "adamic_maybe_boolean", "adamic_method":
		return true
	}
	return false
}

func initializerTokens(source string) []cToken {
	tokens, err := cTokens(source)
	if err != nil {
		panic(err)
	}
	return tokens
}

// Only locals used in another chunk (or final scope cleanup) become address
// parameters. Storage stays in the module wrapper, or in main when final cleanup
// or another module needs it, without a new retain or release.
func (e *emitter) outlineInitializers(parts []initializerPart, cleanup string) string {
	var chunks []initializerChunk
	var modules [][]int
	for partIndex, part := range parts {
		var indexes []int
		count := 0
		for _, statement := range part.statements {
			for _, piece := range initializerStatements(statement) {
				if strings.TrimSpace(piece) == "" {
					continue
				}
				if len(indexes) == 0 || count >= initializerChunkStatements {
					name := stableName("adamic_initialize", part.module, fmt.Sprintf("%q/chunk/%d", part.module, len(indexes)))
					chunks = append(chunks, initializerChunk{module: part.module, name: name, part: partIndex})
					indexes = append(indexes, len(chunks)-1)
					count = 0
				}
				chunks[indexes[len(indexes)-1]].body += piece
				count++
			}
		}
		if len(indexes) > 1 {
			for chunkIndex, index := range indexes {
				chunks[index].name = stableName("adamic_initialize_chunk", part.module, fmt.Sprintf("%q/chunk/%d", part.module, chunkIndex))
			}
		}
		modules = append(modules, indexes)
	}
	var locals []initializerLocal
	owner := map[string]int{}
	for index, chunk := range chunks {
		for _, local := range initializerLocals(chunk.body) {
			if _, found := owner[local.name]; found {
				panic("native: repeated initializer local " + local.name)
			}
			owner[local.name] = index
			locals = append(locals, local)
		}
	}
	shared := map[string]bool{}
	external := map[string]bool{}
	for index, chunk := range chunks {
		for _, token := range initializerTokens(chunk.body) {
			if origin, found := owner[token.text]; found && origin != index {
				shared[token.text] = true
				if chunks[origin].part != chunk.part {
					external[token.text] = true
				}
			}
		}
	}
	for _, token := range initializerTokens(cleanup) {
		if _, found := owner[token.text]; found {
			shared[token.text] = true
			external[token.text] = true
		}
	}
	types := map[string]string{}
	for _, local := range locals {
		if shared[local.name] {
			types[local.name] = local.of
			if external[local.name] {
				e.line("%s %s;", local.of, local.name)
			}
		}
	}
	var definitions strings.Builder
	for index := range chunks {
		chunk := &chunks[index]
		used := map[string]bool{}
		for _, token := range initializerTokens(chunk.body) {
			if shared[token.text] {
				used[token.text] = true
			}
		}
		for _, local := range locals {
			if used[local.name] {
				chunk.parameters = append(chunk.parameters, local.name)
			}
		}
		declarations := initializerLocals(chunk.body)
		removed := map[int]initializerLocal{}
		for _, local := range declarations {
			if shared[local.name] {
				removed[local.start] = local
			}
		}
		var body strings.Builder
		position := 0
		for _, token := range initializerTokens(chunk.body) {
			if token.start < position {
				continue
			}
			body.WriteString(chunk.body[position:token.start])
			if local, found := removed[token.start]; found {
				if local.initialized {
					body.WriteString("(*" + local.name + ") =")
				}
				position = local.end
				continue
			}
			if shared[token.text] {
				body.WriteString("(*" + token.text + ")")
			} else {
				body.WriteString(token.text)
			}
			position = token.end
		}
		body.WriteString(chunk.body[position:])
		signature := initializerSignature(chunk.name, chunk.parameters, types)
		fmt.Fprintf(&definitions, "static %s%s;\n// adamic-module %q\nstatic %s {\n%s\n}\n\n", signature, initializerAttribute, chunk.module, signature, body.String())
	}
	for partIndex, indexes := range modules {
		if len(indexes) == 0 {
			continue
		}
		module := parts[partIndex].module
		used := map[string]bool{}
		for _, index := range indexes {
			for _, parameter := range chunks[index].parameters {
				used[parameter] = true
			}
		}
		var parameters []string
		for _, local := range locals {
			if used[local.name] && external[local.name] {
				parameters = append(parameters, local.name)
			}
		}
		name := stableName("adamic_initialize", module, fmt.Sprintf("%q/module", module))
		signature := initializerSignature(name, parameters, types)
		fmt.Fprintf(&definitions, "static %s%s;\n// adamic-module %q\nstatic %s {\n", signature, initializerAttribute, module, signature)
		// Typed backing arrays avoid thousands of separately described stack
		// variables in a table initializer's wrapper. Chunks still receive typed
		// addresses; the slots have exactly the original storage and lifetime.
		var storageTypes []string
		storageCounts := map[string]int{}
		storage := map[string]string{}
		for _, local := range locals {
			if !used[local.name] || external[local.name] {
				continue
			}
			if storageCounts[local.of] == 0 {
				storageTypes = append(storageTypes, local.of)
			}
			name := stableName("adamic_init_storage", local.of, module+"/"+local.of)
			storage[local.name] = fmt.Sprintf("%s[%d]", name, storageCounts[local.of])
			storageCounts[local.of]++
		}
		for _, of := range storageTypes {
			name := stableName("adamic_init_storage", of, module+"/"+of)
			fmt.Fprintf(&definitions, "\t%s %s[%d];\n", of, name, storageCounts[of])
		}
		for _, index := range indexes {
			var arguments []string
			for _, parameter := range chunks[index].parameters {
				if external[parameter] {
					arguments = append(arguments, parameter)
				} else {
					arguments = append(arguments, "&"+storage[parameter])
				}
			}
			fmt.Fprintf(&definitions, "\t%s(%s);\n", chunks[index].name, strings.Join(arguments, ", "))
		}
		definitions.WriteString("}\n\n")
		e.line("%s(%s);", name, initializerArguments(parameters, true))
	}
	e.out.WriteString(cleanup)
	return definitions.String()
}
func initializerSignature(name string, parameters []string, types map[string]string) string {
	var declarations []string
	for _, parameter := range parameters {
		declarations = append(declarations, types[parameter]+" *"+parameter)
	}
	if len(declarations) == 0 {
		declarations = append(declarations, "void")
	}
	return "void " + name + "(" + strings.Join(declarations, ", ") + ")"
}
func initializerArguments(parameters []string, address bool) string {
	var arguments []string
	for _, parameter := range parameters {
		if address {
			arguments = append(arguments, "&"+parameter)
		} else {
			arguments = append(arguments, parameter)
		}
	}
	return strings.Join(arguments, ", ")
}
