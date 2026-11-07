package lower

import (
	"slices"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// phantomParts separates a primitive from the object views that give it a checker-only name.
// Void is included because tsc's __String deliberately brands undefined through void.
func (lowering *lowering) phantomParts(proven *checker.Type) (*checker.Type, []*checker.Type) {
	proven = lowering.concrete(proven)
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

func (lowering *lowering) phantomBase(proven *checker.Type) *checker.Type {
	primitive, objects := lowering.phantomParts(proven)
	if primitive == nil {
		return nil
	}
	for _, object := range objects {
		if len(lowering.checker.GetSignaturesOfType(object, checker.SignatureKindCall)) != 0 ||
			len(lowering.checker.GetSignaturesOfType(object, checker.SignatureKindConstruct)) != 0 ||
			len(lowering.checker.GetIndexInfosOfType(object)) != 0 {
			return nil
		}
		for _, field := range lowering.checker.GetPropertiesOfType(object) {
			if !lowering.phantomBrandMember(field) || primitiveMember(primitive.Flags(), field.Name) {
				return nil
			}
		}
	}
	return primitive
}

// phantomBrandMember is the one test of a phantom brand member, for a primitive's brand (phantomBase) and
// an object's (brandMintOf, a brand mint's cast) alike (#yth65gz, one proof): keyed by a name, never a
// symbol (Nov 1), and typed so a read finds nothing there (phantomField).
func (lowering *lowering) phantomBrandMember(field *ast.Symbol) bool {
	return !checker.IsKnownSymbol(field) && phantomField(lowering.checker.GetTypeOfSymbol(field))
}

// On a primitive the brand member is never present, so a read gives undefined on Node and natively.
// A required undefined says that as truly as void does, and an optional member's type holds undefined.
func phantomField(proven *checker.Type) bool {
	flags := proven.Flags()
	if flags&checker.TypeFlagsUnion != 0 {
		for _, part := range proven.Types() {
			if !phantomField(part) {
				return false
			}
		}
		return true
	}
	return flags&(checker.TypeFlagsVoid|checker.TypeFlagsUndefined) != 0
}

// These inventories are the own names of boxed primitives and their complete prototype chains,
// observed on Node and independently checked in TestPhantomPrimitiveNames. The checker library is
// intentionally smaller than JavaScript's, so its declarations alone cannot prove absence.
var phantomObjectNames = strings.Fields(
	"__defineGetter__ __defineSetter__ __lookupGetter__ __lookupSetter__ __proto__ constructor hasOwnProperty isPrototypeOf propertyIsEnumerable toLocaleString toString valueOf",
)
var phantomStringNames = strings.Fields(
	"anchor at big blink bold charAt charCodeAt codePointAt concat endsWith fixed fontcolor fontsize includes indexOf isWellFormed italics lastIndexOf length link localeCompare match matchAll normalize padEnd padStart repeat replace replaceAll search slice small split startsWith strike sub substr substring sup toLocaleLowerCase toLocaleUpperCase toLowerCase toUpperCase toWellFormed trim trimEnd trimLeft trimRight trimStart",
)
var phantomNumberNames = strings.Fields("toExponential toFixed toPrecision")

// A symbol key's checker name is internal (a well-known or unique symbol), so no inventory above
// can hold it, and a primitive's prototype has symbol members (String.prototype[Symbol.iterator]).
// Both callers refuse it beside primitiveMember, by tsc's own late-bound name test, whatever the primitive.
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
		index, parseError := strconv.ParseUint(name, 10, 64)
		return parseError == nil && index < 1<<53 && strconv.FormatUint(index, 10) == name
	}
	return flags&checker.TypeFlagsNumberLike != 0 && slices.Contains(phantomNumberNames, name)
}

func (lowering *lowering) phantomUndefined(proven *checker.Type) bool {
	if base := lowering.phantomBase(proven); base != nil {
		return base.Flags()&(checker.TypeFlagsVoid|checker.TypeFlagsUndefined) != 0
	}
	return false
}

// Branding may normalize a read contract without authorizing ordinary branded
// function results or erasure. Those retain the base's separate representation proof.
func (lowering *lowering) viewRepresentation(target *checker.Type) (ir.Type, bool) {
	if base := lowering.phantomBase(target); base != nil {
		return lowering.representation(base)
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		members := make([]*checker.Type, 0, len(target.Types()))
		for _, member := range target.Types() {
			if base := lowering.phantomBase(member); base != nil {
				member = base
			}
			members = append(members, member)
		}
		return lowering.representation(lowering.checker.GetUnionType(members))
	}
	return lowering.representation(target)
}
