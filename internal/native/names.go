package native

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

// stableName always hashes the full identity. Lossy escaping and truncation can
// therefore never make two readable prefixes depend on encounter order.
// Leave room below C11's 127 significant internal-identifier characters for suffixes.
func stableName(prefix, readable, identity string) string {
	digest := sha256.Sum256([]byte(identity))
	text := cIdentifier.ReplaceAllString(readable, "_")
	if len(text) > 60 {
		text = text[:20] + "_" + text[len(text)-39:]
	}
	return prefix + "_" + text + "_" + hex.EncodeToString(digest[:16])
}

func sourceKey(source ir.SourceIdentity) string {
	return strconv.Quote(source.Module) +
		"/" +
		fmt.Sprintf("%q", source.Declaration) +
		"/" +
		strconv.Quote(source.Specialization) +
		"/" +
		strconv.Quote(source.Role)
}

func (nativeEmitter *emitter) functionKey(index int, visiting map[int]bool) string {
	function := nativeEmitter.program.Functions[index]
	if function.Source.Module != "" {
		return sourceKey(function.Source)
	}
	// Compiler-generated helpers have no source declaration. Their semantic IR,
	// with references resolved to names, distinguishes them without program IDs.
	key := function.Name + fmt.Sprint(function.Closure, function.Returns)
	for _, parameter := range function.Parameters {
		key += fmt.Sprint(nativeEmitter.program.Locals[parameter].Type)
	}
	if visiting[index] {
		return key
	}
	visiting[index] = true
	key += nativeEmitter.semanticKey(reflect.ValueOf(function.Body), visiting)
	delete(visiting, index)
	return key
}

func (nativeEmitter *emitter) semanticKey(value reflect.Value, visiting map[int]bool) string {
	if !value.IsValid() {
		return "nil"
	}
	switch value.Kind() {
	case reflect.Interface, reflect.Pointer:
		if value.IsNil() {
			return "nil"
		}
		return nativeEmitter.semanticKey(value.Elem(), visiting)
	case reflect.Slice:
		var out strings.Builder
		out.WriteString(value.Type().String())
		for index := 0; index < value.Len(); index++ {
			out.WriteString("[")
			out.WriteString(nativeEmitter.semanticKey(value.Index(index), visiting))
			out.WriteString("]")
		}
		return out.String()
	case reflect.Struct:
		var out strings.Builder
		out.WriteString(value.Type().String())
		for index := 0; index < value.NumField(); index++ {
			field := value.Type().Field(index).Name
			child := value.Field(index)
			out.WriteString("/" + field + "=")
			if child.Kind() == reflect.Int {
				index := int(child.Int())
				switch field {
				case "Function":
					if index >= 0 && index < len(nativeEmitter.program.Functions) {
						out.WriteString(nativeEmitter.functionKey(index, visiting))
						continue
					}
				case "Site":
					out.WriteString("allocation")
					continue
				case "Class":
					if index > 0 {
						if value.Type() == reflect.TypeOf(ir.Property{}) ||
							value.Type() == reflect.TypeOf(ir.SetProperty{}) {
							out.WriteString(nativeEmitter.functionKey(index-1, visiting))
						} else {
							out.WriteString(nativeEmitter.classKey(index, visiting))
						}
						continue
					}
				case "Local", "CatchLocal":
					if index >= 0 && index < len(nativeEmitter.program.Locals) {
						local := nativeEmitter.program.Locals[index]
						out.WriteString(local.Name + fmt.Sprint(local.Type))
						continue
					}
				case "Source", "Flags":
					if value.Type() == reflect.TypeOf(ir.RegExpNew{}) {
						out.WriteString(strconv.Quote(nativeEmitter.program.Strings[index]))
						continue
					}
				case "Index":
					if value.Type() == reflect.TypeOf(ir.RegExpNew{}) {
						out.WriteString(nativeEmitter.regexName(index))
						continue
					}
					if value.Type() == reflect.TypeOf(ir.StringConstant{}) {
						out.WriteString(strconv.Quote(nativeEmitter.program.Strings[index]))
						continue
					}
				}
			}
			out.WriteString(nativeEmitter.semanticKey(child, visiting))
		}
		return out.String()
	default:
		return fmt.Sprint(value.Interface())
	}
}

func (nativeEmitter *emitter) sourceReadable(index int) string {
	function := nativeEmitter.program.Functions[index]
	if function.Source.Module != "" {
		return function.Source.Module + "_" + strings.Join(function.Source.Declaration, "_")
	}
	return function.Name
}
func (nativeEmitter *emitter) namedFunction(index int) string {
	return stableName(
		"adamic_function",
		nativeEmitter.sourceReadable(index),
		nativeEmitter.functionKey(index, map[int]bool{}),
	)
}

func (nativeEmitter *emitter) namedLocal(index int) string {
	if nativeEmitter.localNames == nil {
		nativeEmitter.localNames = map[int]string{}
		occurrences := map[string]int{}
		for index, local := range nativeEmitter.program.Locals {
			prefix, owner := "adamic_local", local.Source.Module
			if local.Global {
				prefix = "adamic_global"
			} else if local.Function >= 0 && local.Function < len(nativeEmitter.program.Functions) {
				owner = nativeEmitter.functionName(local.Function)
			}
			module := ""
			if local.Global {
				module = local.Source.Module
			}
			group := fmt.Sprintf("%t/%d/%q/%q", local.Global, local.Function, module, local.Name)
			occurrence := occurrences[group]
			occurrences[group]++
			key := owner + "/" + local.Name + "/" + strconv.Itoa(occurrence)
			if local.Global && len(local.Source.Declaration) > 0 {
				key = sourceKey(local.Source) + "/" + local.Name
			}
			digest := sha256.Sum256([]byte(key))
			readable := cIdentifier.ReplaceAllString(local.Name, "_")
			if len(readable) > 60 {
				readable = readable[:60]
			}
			// Preserve the readable trailing spelling; these digits are a digest, never an ordinal.
			nativeEmitter.localNames[index] = prefix +
				"_" +
				new(big.Int).SetBytes(digest[:16]).String() +
				"_" +
				readable
		}
	}
	return nativeEmitter.localNames[index]
}

func stringName(value string) string { return stableName("adamic_string", "literal", value) }

func (nativeEmitter *emitter) classKey(index int, visiting map[int]bool) string {
	class := nativeEmitter.program.Classes[index-1]
	key := sourceKey(class.Source) + "/" + class.Name
	if class.Literal {
		for _, accessor := range class.Accessors {
			key += "/" + strconv.Quote(accessor.Name)
			if accessor.Getter >= 0 {
				key += "/get=" + nativeEmitter.functionKey(accessor.Getter, visiting)
			}
			if accessor.Setter >= 0 {
				key += "/set=" + nativeEmitter.functionKey(accessor.Setter, visiting)
			}
		}
	}
	if class.Constructor >= 0 {
		key = nativeEmitter.functionKey(class.Constructor, visiting)
	}
	return key
}

func (nativeEmitter *emitter) className(index int) string {
	return stableName(
		"adamic_class",
		nativeEmitter.program.Classes[index-1].Name,
		nativeEmitter.classKey(index, map[int]bool{}),
	)
}

// sharedDeclaration deduplicates identical helper declarations made at several sites.
func (nativeEmitter *emitter) sharedDeclaration(prefix, key string) (string, bool) {
	name := stableName(prefix, "data", key)
	if nativeEmitter.generatedDeclarations == nil {
		nativeEmitter.generatedDeclarations = map[string]bool{}
	}
	found := nativeEmitter.generatedDeclarations[name]
	nativeEmitter.generatedDeclarations[name] = true
	return name, found
}

// accessorStorageName is a private implementation key, never a public property name.
func (nativeEmitter *emitter) accessorStorageName(function int) string {
	return "#accessor:" + nativeEmitter.functionName(function)
}
func (nativeEmitter *emitter) printedFieldName(field ir.Field) string {
	if field.Private && strings.HasPrefix(field.Name, "#accessor:") {
		index, parseError := strconv.Atoi(strings.TrimPrefix(field.Name, "#accessor:"))
		if parseError != nil || index < 0 || index >= len(nativeEmitter.program.Functions) {
			panic("native: invalid accessor storage identity")
		}
		return nativeEmitter.accessorStorageName(index)
	}
	return field.Name
}
