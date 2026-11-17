# Representation reason mutants

Each Go overlay blanks one reason while preserving the IR type and known flag.
Each focused `go test -count=1 -p=1 -vet=off -overlay … ./internal/lower/ -run …`
exited 1 at the listed assertion. The overlays leave the working tree intact.

| Mutant | Changed return | Red test and assertion |
| --- | --- | --- |
| Blank array member reason | `expression.go:249` | `TestRepresentationReasonStops/array_member` at `representation_reason_test.go:72` |
| Blank index signature reason | `expression.go:253` | `TestRepresentationReasonStops/index_signature` at `representation_reason_test.go:75` |
| Blank intersection member reason | `library_object.go:341` | `TestRepresentationReasonStops/intersection_array` at `representation_reason_test.go:75` |
| Blank intersection field reason | `library_object.go:363` | `TestRepresentationReasonStops/optional_boolean` at `representation_reason_test.go:75` |
| Blank nullable union member reason | `expression.go:173` | `TestRepresentationReasonStops/nullable_union_member` at `representation_reason_test.go:75` |
| Blank union member reason | `expression.go:197` | `TestRepresentationReasonStops/union_member` at `representation_reason_test.go:75` |
| Blank weak union reason | `expression.go:206` | `TestRepresentationReasonStops/weak_union` at `representation_reason_test.go:75` |
| Blank scalar undefined reason | `expression.go:219` | `TestRepresentationReasonScalarUndefined` at `representation_reason_test.go:99` |
| Blank phantom union reason | `expression.go:222` | `TestRepresentationReasonStops/phantom_union` at `representation_reason_test.go:75` |
| Blank type parameter reason | `expression.go:120` | `TestRepresentationReasonStops/type_parameter` at `representation_reason_test.go:75` |
| Blank never reason | `expression.go:228` | `TestRepresentationReasonStops/never` at `representation_reason_test.go:75` |
| Blank symbol reason | `expression.go:230` | `TestRepresentationReasonStops/symbol` at `representation_reason_test.go:75` |
| Blank bigint reason | `expression.go:232` | `TestRepresentationReasonStops/bigint` at `representation_reason_test.go:75` |
| Blank indexed access reason | `expression.go:234` | `TestRepresentationReasonStops/indexed_access` at `representation_reason_test.go:75` |
| Blank conditional reason | `expression.go:236` | `TestRepresentationReasonStops/conditional` at `representation_reason_test.go:75` |
| Blank keyof reason | `expression.go:238` | `TestRepresentationReasonStops/keyof` at `representation_reason_test.go:75` |
| Blank fallback reason | `expression.go:240` | `TestRepresentationReasonStops/fallback` at `representation_reason_test.go:61` |
