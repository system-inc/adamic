Mutant evidence for the form-scoped require row in docs/axioms/host-any.md
(module.d.ts:707, #3sh9398 and Adamic's #redxes2 ruling).

Mutant: widen nodeCommonJSRequireThrows in library_node_commonjs.go to accept
"fs" beside nodeSourceMapSupport. The Go overlay changes only that comparison.
Command:

```sh
go test -json -p 1 ./internal/lower -count=1 -timeout=5m \
  -overlay=/tmp/host-any-require-evidence/fs-overlay.json \
  -run '^TestHostAnyModelledRequireForms$'
```

Result with the mutant: exit 1. Restored without the overlay, the test passes.
The real failing line (the no-substitution template probe fails too):

```text
host_any_require_modelled_test.go:75: nodeCommonJSRequireThrows(require("fs")) = true; Modelled Forms require false
```

Second mutant: delete the Modelled row for require from host-any.md.
Command:

```sh
go test -json -p 1 ./internal/lower -count=1 -timeout=5m \
  -run '^TestHostAnyModelledRequireForms$'
```

Result with the row deleted: exit 1. Restored, the same command passes.
The real failing line:

```text
host_any_require_modelled_test.go:33: host-any.md must name exactly the two Modelled require forms at module.d.ts:707; got 0 rows, forms map[]
```
