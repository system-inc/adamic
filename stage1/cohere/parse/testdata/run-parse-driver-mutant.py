#!/usr/bin/env python3
"""Both driver backends report the same wrong AST; only Go detects it."""
import pathlib
import subprocess

root = pathlib.Path(__file__).resolve().parents[4]
source = root / "stage1/cohere/parse/parse.a"
original = source.read_text()
before = "    context.root = root;"
after = "    const rootNode = parser.nodes[root] ?? panic('missing root');\n    rootNode.semantic = 'mutant';\n" + before
assert original.count(before) == 1
try:
    source.write_text(original.replace(before, after))
    result = subprocess.run(["go", "test", "./stage1/cohere/parse", "-run", "^TestParseDriverAST$", "-count=1", "-timeout", "55s", "-v"], cwd=root, text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, timeout=180)
    print(result.stdout, end="")
    assert result.returncode != 0, "AST field mutant survived"
    assert "full AST differs from Go: native/Node agree=true" in result.stdout, "mutant must fail only at the Go comparison"
    print("RED: AST root semantic mutant caught only by TestParseDriverAST's Go oracle")
finally:
    source.write_text(original)
