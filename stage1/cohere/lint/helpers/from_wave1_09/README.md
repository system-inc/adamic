# Wave 09 helper candidates

`nodes_from_static_declarations.a` ports the actual Go helper. Import `nodesFromStaticDeclarations` and the `StaticDeclaration`/`DeclarationNode` types. The returned list and every node are fresh, ordered, and independent of input mutation. `contextPresent` and `nodesPresent` retain the nil container fields of the Go declaration shape; this declaration-only boundary does not provide container access.

Run the owned comparator with the setup environment sourced:

```sh
go test ./stage1/cohere/lint/helpers/from_wave1_09 -count=1 -v -timeout=15m > /tmp/lint-helpers-wave109-test.log 2>&1
```

See [REPORT.md](REPORT.md) for coverage, mutants and integration limits.

`property_sort.a` exports `propertySort(arena, roots, propertyOrder)`. It returns `order`, `count` and the exact declaration `visited` trace. Arena nodes have indexed children. Only rule/at-rule children are enqueued. Provide the real PropertyOrder map and a finite acyclic tree; missing indices refuse rather than silently skip work.

`parse_value.a` exports `parseValue(source, separators)`, returning an acyclic `arena` and ordered `roots`. Each node retains exact text, kind, indexed children and `nodesPresent`. Pass the membership set from actual Go isValueSeparator; no CSS parser or serializer integration is implied.

The bounded helper gate passes; required original live-consumer validation is blocked by missing Tailwind and Kirk-local corpus files. These are candidates, with potential readiness credit only. See the final section of REPORT.md and evidence/live-consumers/. The next loader was withdrawn after proving a lossy shared file-input boundary.
