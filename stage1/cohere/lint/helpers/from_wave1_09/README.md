# Static declaration nodes

`nodes_from_static_declarations.a` ports the actual Go helper. Import `nodesFromStaticDeclarations` and the `StaticDeclaration`/`DeclarationNode` types. The returned list and every node are fresh, ordered, and independent of input mutation. `contextPresent` and `nodesPresent` retain the nil container fields of the Go declaration shape; this declaration-only boundary does not provide container access.

Run the owned comparator with the setup environment sourced:

```sh
go test ./stage1/cohere/lint/helpers/from_wave1_09 -count=1 -v -timeout=15m > /tmp/lint-helpers-wave109-test.log 2>&1
```

See [REPORT.md](REPORT.md) for coverage, mutants and integration limits.
