# Overlay drift validation

Base: e50e502c21f12f9df64f7e716e980ed84695aa6f, fetched origin/main on October 23.
No production compiler source changed.

`make_overlay.py` removes the body of `Lower`, including `flow.NormalizeAsync`,
`flow.AsyncError` and `errors.As`. That removed the only uses of flow and errors.
The final AST import pass now cleans every generated overlay file, including
files extended after their first write. Blank and dot imports retain their semantics.

The preflight matcher checks top-level if/short-declaration/call/condition/return
AST structure. Its selector name is unrestricted. Existing visitor and outer-return
checks remain. The slice-3 snapshot includes mergedFieldRulings and functionAnnotations;
the compiled continuation fixture adds futurePreflight. Wrapped returns, nested
preflights, assignment initializers and non-module arguments are refused.

Commands (all output redirected to the corresponding compressed logs here):

```sh
export GOPROXY='https://proxy.golang.org|direct'
timeout 900 bash cloud/setup.sh > /tmp/ck10df5-setup.log 2>&1
source /workspace/adamic-tools/env.sh
timeout 360 python3 stage3/census/latent/audit_overlay_drift.py /tmp/overlay-drift-mutants > /tmp/ck10df5-mutants.log 2>&1
timeout 240 go test ./stage3/census/latent/refusalrewrite/... ./stage3/census/latent/importclean ./stage3/census/latent/statecopy ./stage3/census/latent/statementrewrite -run '^TestProduct_' -count=1 > /tmp/ck10df5-products.log 2>&1
timeout 240 go test ./stage3/census/latent/refusalrewrite/... ./stage3/census/latent/importclean ./stage3/census/latent/statecopy ./stage3/census/latent/statementrewrite -skip '^TestProduct_' -count=1 > /tmp/ck10df5-tests.log 2>&1
```

Both test commands exited 0. The mutant runner exited 0 only after both child tests
exited 1 with their intended diagnostic. Name-list restoration failed
TestCompilerSlice3Shape with unsupported outer refusal returns. Returning the
uncleaned rewritten source failed TestProduct_RemovedImportBuilds at Go compilation
with errors imported and not used. Each mutation was restored before the next run.
The import fixture replaces the production body that used errors, then cleans and
builds the resulting file, retaining an explicit fmt alias and a side-effect import.

Setup initially blocked focused testing while the TypeScript submodule was absent;
those attempts did not compile tests. Once dependencies were present, focused tests
and the complete four generator packages passed. Setup timing lines: Node 0.066s,
Go 0.076s, clang 0.499s, Markdown dependencies 1.064s, submodules 272.330s,
Go build 426.920s, build cache 427.032s, done 427.076s. nproc: 5;
cgroup CPU quota: 400000/100000. Node v24.19.0, Go 1.27.1, clang 20.1.8.
