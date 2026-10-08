package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"slices"
	"strconv"
	"strings"
)

// Map keys use the primitive phantom-brand rule from Adamic's d2d3c77e.
// Keep this hook local to key selection until the full brand unit lands. In
// particular, never fields, real properties and callable/indexed brands do not
// become string keys. T is substituted before its proven brand is examined.
func (l *lowering) mapKeyRepresentation(proven *checker.Type) (ir.Type, bool) {
	proven = l.concrete(proven)
	if held, known := l.representation(proven); known {
		return held, true
	}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			held, known := l.mapKeyRepresentation(member)
			if !known || held != ir.String {
				return 0, false
			}
		}
		return ir.String, true
	}
	if proven.Flags()&checker.TypeFlagsIntersection == 0 {
		return 0, false
	}
	var primitive *checker.Type
	var brands []*checker.Type
	for _, member := range proven.Types() {
		switch {
		case member.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsVoid) != 0:
			if primitive != nil {
				return 0, false
			}
			primitive = member
		case member.Flags()&checker.TypeFlagsObject != 0:
			brands = append(brands, member)
		default:
			return 0, false
		}
	}
	if primitive == nil || len(brands) == 0 {
		return 0, false
	}
	for _, brand := range brands {
		if len(l.checker.GetSignaturesOfType(brand, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(brand, checker.SignatureKindConstruct)) != 0 || len(l.checker.GetIndexInfosOfType(brand)) != 0 {
			return 0, false
		}
		for _, field := range l.checker.GetPropertiesOfType(brand) {
			if l.checker.GetTypeOfSymbol(field).Flags()&checker.TypeFlagsVoid == 0 {
				return 0, false
			}
			if primitive.Flags()&checker.TypeFlagsStringLike != 0 && mapStringMember(field.Name) {
				return 0, false
			}
		}
	}
	// A string brand is the string itself. __String's branded void arm is
	// undefined, the NULL string key already supported by the Map runtime.
	return ir.String, true
}

// Node's complete boxed-string own/prototype names, from the phantom-brand
// unit's independently checked inventory, not the smaller checker library.
var mapStringMembers = strings.Fields("__defineGetter__ __defineSetter__ __lookupGetter__ __lookupSetter__ __proto__ constructor hasOwnProperty isPrototypeOf propertyIsEnumerable toLocaleString toString valueOf anchor at big blink bold charAt charCodeAt codePointAt concat endsWith fixed fontcolor fontsize includes indexOf isWellFormed italics lastIndexOf length link localeCompare match matchAll normalize padEnd padStart repeat replace replaceAll search slice small split startsWith strike sub substr substring sup toLocaleLowerCase toLocaleUpperCase toLowerCase toUpperCase toWellFormed trim trimEnd trimLeft trimRight trimStart")

func mapStringMember(name string) bool {
	if slices.Contains(mapStringMembers, name) {
		return true
	}
	index, err := strconv.ParseUint(name, 10, 64)
	return err == nil && index < 1<<53 && strconv.FormatUint(index, 10) == name
}
