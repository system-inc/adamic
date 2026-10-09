package lower

import (
	"slices"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// phantomParts separates a primitive from the object views that give it a checker-only name.
// Void is included because tsc's __String deliberately brands undefined through void.
func (l *lowering) phantomParts(proven *checker.Type) (*checker.Type, []*checker.Type) {
	proven = l.concrete(proven)
	if proven.Flags()&checker.TypeFlagsIntersection == 0 {
		return nil, nil
	}
	var primitive *checker.Type
	var objects []*checker.Type
	for _, part := range proven.Types() {
		switch {
		case part.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsVoid|checker.TypeFlagsUndefined) != 0:
			if primitive != nil {
				return nil, nil
			}
			primitive = part
		case part.Flags()&checker.TypeFlagsObject != 0:
			objects = append(objects, part)
		default:
			return nil, nil
		}
	}
	if primitive == nil || len(objects) == 0 {
		return nil, nil
	}
	return primitive, objects
}

func (l *lowering) phantomBase(proven *checker.Type) *checker.Type {
	primitive, objects := l.phantomParts(proven)
	if primitive == nil {
		return nil
	}
	for _, object := range objects {
		if len(l.checker.GetSignaturesOfType(object, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(object, checker.SignatureKindConstruct)) != 0 || len(l.checker.GetIndexInfosOfType(object)) != 0 {
			return nil
		}
		for _, field := range l.checker.GetPropertiesOfType(object) {
			if !phantomField(l.checker.GetTypeOfSymbol(field), field.Flags&ast.SymbolFlagsOptional != 0) || primitiveMember(primitive.Flags(), field.Name) {
				return nil
			}
		}
	}
	return primitive
}

func phantomField(proven *checker.Type, optional bool) bool {
	flags := proven.Flags()
	if flags&checker.TypeFlagsUnion != 0 {
		for _, part := range proven.Types() {
			if !phantomField(part, optional) {
				return false
			}
		}
		return true
	}
	return flags&checker.TypeFlagsVoid != 0 || optional && flags&checker.TypeFlagsUndefined != 0
}

// These inventories are the own names of boxed primitives and their complete prototype chains,
// observed on Node and independently checked in TestPhantomPrimitiveNames. The checker library is
// intentionally smaller than JavaScript's, so its declarations alone cannot prove absence.
var phantomObjectNames = strings.Fields("__defineGetter__ __defineSetter__ __lookupGetter__ __lookupSetter__ __proto__ constructor hasOwnProperty isPrototypeOf propertyIsEnumerable toLocaleString toString valueOf")
var phantomStringNames = strings.Fields("anchor at big blink bold charAt charCodeAt codePointAt concat endsWith fixed fontcolor fontsize includes indexOf isWellFormed italics lastIndexOf length link localeCompare match matchAll normalize padEnd padStart repeat replace replaceAll search slice small split startsWith strike sub substr substring sup toLocaleLowerCase toLocaleUpperCase toLowerCase toUpperCase toWellFormed trim trimEnd trimLeft trimRight trimStart")
var phantomNumberNames = strings.Fields("toExponential toFixed toPrecision")

func primitiveMember(flags checker.TypeFlags, name string) bool {
	if flags&(checker.TypeFlagsVoid|checker.TypeFlagsUndefined) != 0 {
		return false
	}
	if slices.Contains(phantomObjectNames, name) {
		return true
	}
	if flags&checker.TypeFlagsStringLike != 0 {
		if slices.Contains(phantomStringNames, name) {
			return true
		}
		// A string may have any nonnegative integer index below its length. Noncanonical spellings
		// such as 00 and -0 aren't string exotic own properties.
		index, err := strconv.ParseUint(name, 10, 64)
		return err == nil && index < 1<<53 && strconv.FormatUint(index, 10) == name
	}
	return flags&checker.TypeFlagsNumberLike != 0 && slices.Contains(phantomNumberNames, name)
}

func (l *lowering) phantomUndefined(proven *checker.Type) bool {
	if base := l.phantomBase(proven); base != nil {
		return base.Flags()&(checker.TypeFlagsVoid|checker.TypeFlagsUndefined) != 0
	}
	return false
}
