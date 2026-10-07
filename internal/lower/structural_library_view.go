package lower

import "github.com/microsoft/TypeScript/tsc/shim/checker"

// libraryPrototypeView finds a library prototype member exposed as a structural field
// or own method. Its intrinsic lowering needs the library receiver type; a structural
// view loses that type without putting the member into the native object's shape.
// Follow the same slots as keeping checks, including literal and shorthand fields.
func (l *lowering) libraryPrototypeView(from, to *checker.Type, visited map[[2]*checker.Type]bool) string {
	if from == nil || to == nil || from == to || visited[[2]*checker.Type{from, to}] {
		return ""
	}
	visited[[2]*checker.Type{from, to}] = true
	if from.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range from.Types() {
			if name := l.libraryPrototypeView(member, to, visited); name != "" {
				return name
			}
		}
		return ""
	}
	if to.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range to.Types() {
			if name := l.libraryPrototypeView(from, member, visited); name != "" {
				return name
			}
		}
		return ""
	}
	from, to = l.present(from), l.present(to)
	if from == nil || to == nil || from == to {
		return ""
	}
	if l.librarySymbol(from.Symbol()) && !l.librarySymbol(to.Symbol()) {
		for _, viewed := range l.checker.GetPropertiesOfType(to) {
			member := l.checker.GetPropertyOfType(from, viewed.Name)
			if l.inheritedLibrarySymbol(member) && !l.libraryOwnMember(from, viewed.Name) {
				return viewed.Name
			}
		}
	}
	fromSignatures := l.checker.GetSignaturesOfType(from, checker.SignatureKindCall)
	toSignatures := l.checker.GetSignaturesOfType(to, checker.SignatureKindCall)
	if len(fromSignatures) > 0 && len(toSignatures) > 0 {
		fromParameters, toParameters := fromSignatures[0].Parameters(), toSignatures[0].Parameters()
		for index := 0; index < len(fromParameters) && index < len(toParameters); index++ {
			// A function is called with the view's arguments, and returns its own result.
			if name := l.libraryPrototypeView(l.checker.GetTypeOfSymbol(toParameters[index]), l.checker.GetTypeOfSymbol(fromParameters[index]), visited); name != "" {
				return name
			}
		}
		return l.libraryPrototypeView(l.checker.GetReturnTypeOfSignature(fromSignatures[0]), l.checker.GetReturnTypeOfSignature(toSignatures[0]), visited)
	}
	if from.ObjectFlags()&checker.ObjectFlagsReference != 0 && to.ObjectFlags()&checker.ObjectFlagsReference != 0 && (l.checker.IsArrayType(from) || checker.IsTupleType(from) || l.isLibraryType(from, "Map", "ReadonlyMap", "Set", "ReadonlySet")) {
		fromArguments, toArguments := l.typeArguments(from), l.typeArguments(to)
		for index := 0; index < len(fromArguments) && index < len(toArguments); index++ {
			if name := l.libraryPrototypeView(fromArguments[index], toArguments[index], visited); name != "" {
				return name
			}
		}
		return ""
	}
	for _, viewed := range l.checker.GetPropertiesOfType(to) {
		if inside := l.checker.GetPropertyOfType(from, viewed.Name); inside != nil {
			if name := l.libraryPrototypeView(l.checker.GetTypeOfSymbol(inside), l.checker.GetTypeOfSymbol(viewed), visited); name != "" {
				return name
			}
		}
	}
	return ""
}

// The bundled library declares both own data and prototype members. These own fields
// are not prototype hazards; their individual reads still need a supported representation.
func (l *lowering) libraryOwnMember(proven *checker.Type, name string) bool {
	switch {
	case l.isLibraryType(proven, "RegExp"):
		return name == "lastIndex"
	case l.isLibraryType(proven, "RegExpExecArray", "RegExpMatchArray", "RegExpIndicesArray"):
		return name == "length" || name == "index" || name == "input" || name == "groups" || name == "indices"
	case l.checker.IsArrayType(proven), l.isLibraryType(proven, "String"):
		return name == "length"
	case l.isLibraryType(proven, "Error"):
		return name == "message" || name == "cause" || name == "stack"
	case l.isLibraryType(proven, "IteratorYieldResult", "IteratorReturnResult"):
		return name == "done" || name == "value"
	case l.isLibraryType(proven, "Function"):
		return name == "length" || name == "name"
	}
	return false
}
