"""Scratch, never landing (#p1qw4kf): plant one mutant of splice_spread/mutants.md, one exact line replaced."""

import sys

mutants = {
    "M1": (
        "internal/lower/array_call_arguments.go",
        "\tat := ir.Binary{Operator: ir.Add, Left: b.read(index), Right: b.read(position)}\n",
        "\tat := ir.Binary{Operator: ir.Add, Left: start, Right: b.read(position)}\n",
    ),
    "M2": (
        "internal/lower/array_call_arguments.go",
        "\tb.body = append(b.body, removal, insertion)\n",
        "\tb.body = append(b.body, insertion, removal)\n",
    ),
    "M3": (
        "internal/lower/array_call_arguments.go",
        "\tb := l.libraryArrayBuilder([]ir.Expression{array, arguments[0], arguments[1], packed})\n",
        "\tunpacked := ir.Expression(packed)\n"
        "\tif len(packed.Elements) == 1 && packed.Spread[0] {\n"
        "\t\tunpacked = packed.Elements[0]\n"
        "\t}\n"
        "\tb := l.libraryArrayBuilder([]ir.Expression{array, arguments[0], arguments[1], unpacked})\n",
    ),
    "M4": (
        "internal/lower/object.go",
        '\tif name == "splice" && arraySpliceSpreads(node.AsCallExpression().Arguments.Nodes) {\n',
        '\tif false && name == "splice" && arraySpliceSpreads(node.AsCallExpression().Arguments.Nodes) {\n',
    ),
}

name = sys.argv[1]
path, old, new = mutants[name]
with open(path) as file:
    lines = file.readlines()
found = [index for index, line in enumerate(lines) if line == old]
if len(found) != 1:
    print(f"mutant {name}: {len(found)} lines match in {path}, want 1")
    sys.exit(1)
lines[found[0]] = new
with open(path, "w") as file:
    file.writelines(lines)
print(f"mutant {name}: {path}:{found[0] + 1} replaced")
