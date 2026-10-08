# Core helpers

`testdata/helpers.json` lists the 17 exact Go symbols, implementation paths and captured call counts. `nodes.a` preserves projected AST facts and node identity; `from_parser.a` supplies the token-signature facts needed by the delivered rule. Other consumers must supply their named AST facts.

The Go overlay in `testdata/capture.py` observes every helper invocation in the consuming upstream suites, including recursive calls. It never edits cohere. With the setup environment sourced, regenerate using `python3 stage1/cohere/lint/helpers/core/testdata/capture.py`, then run `go test ./stage1/cohere/lint/helpers/... -count=1 -v -timeout=20m`. The compressed corpus contains 8,230 calls from 2,552 frames at cohere pin `7945d102a6c18dd36adf9114a758ce646e8b2359`. Replay compares exact result bytes on Node, emitted JavaScript and ASan/UBSan native; each of the 17 mutants must disagree with Go on all three.

No partial core port existed to reuse. `codePathRoots` is stopped at `cohere/internal/lint/rules/core/code_path_roots.go:17`: its dependency `ecmascript/control_flow_graph.IndexRoots` remains unported.

The consuming proof is `stage1/cohere/lint/rules/no-dupe-else-if/`, with 35 unique upstream cases (70 runs). Its `testdata/native_mutant.py` uses a temporary Go test overlay to select this rule as the sanitized native canary, while leaving the shared harness unchanged.
