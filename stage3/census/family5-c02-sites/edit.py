"""Apply the measured annotations to a copy of the fully adapted tree."""

import sys
from pathlib import Path


def update(path, edits):
    original = path.read_bytes()
    updated = original
    for before, after in edits:
        before, after = before.encode(), after.encode()
        assert updated.count(before) == 1, (path, before)
        updated = updated.replace(before, after)
    path.write_bytes(updated)


compiler = Path(sys.argv[1]) / "src/compiler"
line_view = "LineMapSource"
update(compiler / "scanner.ts", [
    ("getLinesBetweenPositions(sourceFile: SourceFileLike,", f"getLinesBetweenPositions(sourceFile: {line_view},"),
])
update(compiler / "utilities.ts", [
    ("    type SourceFileLike,", "    type LineMapSource, type SourceFileLike,"),
    ("getEndLinePosition(line: number, sourceFile: SourceFileLike)", f"getEndLinePosition(line: number, sourceFile: {line_view})"),
    ("getTokenPosOfNode(node: Node, sourceFile?: SourceFileLike,", 'getTokenPosOfNode(node: Node, sourceFile?: { readonly text: string },'),
])
update(compiler / "factory/nodeChildren.ts", [
    ("    type SourceFileLike,\r\n", ""),
    ("new WeakMap<SourceFileLike,", "new WeakMap<object,"),
    ("getNodeChildren(node: Node, sourceFile: SourceFileLike)", "getNodeChildren(node: Node, sourceFile: object)"),
    ("setNodeChildren(node: Node, sourceFile: SourceFileLike,", "setNodeChildren(node: Node, sourceFile: object,"),
    ("unsetNodeChildren(node: Node, origSourceFile: SourceFileLike)", "unsetNodeChildren(node: Node, origSourceFile: object)"),
    ("transferSourceFileChildren(sourceFile: SourceFileLike, targetSourceFile: SourceFileLike)", "transferSourceFileChildren(sourceFile: object, targetSourceFile: object)"),
])
