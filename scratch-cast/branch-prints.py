import sys

# The branch probe's temporary prints, into castRoad of whichever cast_proof.go is in place (main's or the head's),
# on stderr, each line opening BRANCH. Scratch only: never in the code commit.
path = sys.argv[1]
text = open(path).read()


def insert(anchor, lines, before=False):
    global text
    if text.count(anchor) != 1:
        print(f"FAIL branch prints: {anchor!r} occurs {text.count(anchor)} times")
        sys.exit(1)
    text = text.replace(anchor, lines + anchor if before else anchor + lines)


insert('import (\n', '\t"fmt"\n\t"os"\n')
insert(
    'func (l *lowering) castRoad(node *ast.Node, source, target *checker.Type, stop func(castCause, string) error) (castProof, error) {\n',
    '\tfmt.Fprintf(os.Stderr, "BRANCH castRoad: source %q (object %v, intersection %v, union %v), target %q (object %v)\\n", l.checker.TypeToString(source), source.Flags()&checker.TypeFlagsObject != 0, source.Flags()&checker.TypeFlagsIntersection != 0, source.Flags()&checker.TypeFlagsUnion != 0, l.checker.TypeToString(target), target.Flags()&checker.TypeFlagsObject != 0)\n',
)
insert(
    '\t\tif cast, err := l.kindUnionCast(node, source, target); cast != nil || err != nil {\n',
    '\t\tfmt.Fprintf(os.Stderr, "BRANCH interfaceCast: no road and no stop\\n")\n',
    before=True,
)
insert(
    '\t\tto, toKnown := l.representation(target)\n',
    '\t\tfmt.Fprintf(os.Stderr, "BRANCH kindUnionCast: no road and no stop\\n")\n'
    '\t\tfmt.Fprintf(os.Stderr, "BRANCH structural view asks: source object %v, target object %v, from known %v and Object %v, to known %v and Object %v, target a class %v, target assignable to source %v, index signatures %d\\n", source.Flags()&checker.TypeFlagsObject != 0, target.Flags()&checker.TypeFlagsObject != 0, fromKnown, from == ir.Object, toKnown, to == ir.Object, isClassInstance(target), l.checker.IsTypeAssignableTo(target, source), len(l.checker.GetIndexInfosOfType(target)))\n'
    '\t\tif found := l.widened(target, source, map[[2]*checker.Type]bool{}); found != nil {\n'
    '\t\t\tfmt.Fprintf(os.Stderr, "BRANCH widened: %q written where %q is read, readonly field %q\\n", l.checker.TypeToString(found.target), l.checker.TypeToString(found.source), found.readonlyField)\n'
    '\t\t} else {\n'
    '\t\t\tfmt.Fprintf(os.Stderr, "BRANCH widened: nothing\\n")\n'
    '\t\t}\n'
    '\t\tfor _, property := range l.checker.GetPropertiesOfType(target) {\n'
    '\t\t\tif inside := l.checker.GetPropertyOfType(source, property.Name); inside != nil {\n'
    '\t\t\t\tfmt.Fprintf(os.Stderr, "BRANCH field %s: target %q readonly %v, source %q readonly %v, target to source %v, source to target %v\\n", property.Name, l.checker.TypeToString(l.checker.GetTypeOfSymbol(property)), l.checker.IsReadonlySymbol(property), l.checker.TypeToString(l.checker.GetTypeOfSymbol(inside)), l.checker.IsReadonlySymbol(inside), l.checker.IsTypeAssignableTo(l.checker.GetTypeOfSymbol(property), l.checker.GetTypeOfSymbol(inside)), l.checker.IsTypeAssignableTo(l.checker.GetTypeOfSymbol(inside), l.checker.GetTypeOfSymbol(property)))\n'
    '\t\t\t}\n'
    '\t\t}\n',
)
insert(
    '\t\treturn castProof{}, stop(cause, detail)\n',
    '\t\tfmt.Fprintf(os.Stderr, "BRANCH objectCastCause: cause %d, castUnnamed %v, castWidenedTarget %v, castSideways %v, detail %q\\n", cause, cause == castUnnamed, cause == castWidenedTarget, cause == castSideways, detail)\n',
    before=True,
)
open(path, 'w').write(text)
print("branch prints in " + path)
