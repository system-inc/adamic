# narrowedNeverReadGap's mutants (#v59xyph)

Each is one edit to internal/lower/narrowed_nullish.go, applied alone, then the named test run with -v.

## 1. The rule unregistered

Old:

```go
	refusalRules.register(3340, "narrowedNeverReadGap", (*lowering).narrowedNeverReadGap)
```

New:

```go
	_ = (*lowering).narrowedNeverReadGap
```

Expected: TestNeverReadOfACallWrittenLetIsNotYet goes red on every row, each lowering with no error, and
TestRefusalRulesRunInOrder misses 3340. Built natively under --sanitize, nv9 is the AddressSanitizer
SEGV in adamic_string_to_upper again, where Node throws.

## 2. The writtenAcrossFunctions gate dropped

Old:

```go
	if symbol == nil || symbol.Flags&ast.SymbolFlagsVariable == 0 || !l.writtenAcrossFunctions(symbol) {
		return nil
	}
	if l.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsNever == 0 {
```

New:

```go
	if symbol == nil || symbol.Flags&ast.SymbolFlagsVariable == 0 {
		return nil
	}
	if l.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsNever == 0 {
```

Expected: TestNeverReadOfAnUnwrittenParameterLowers goes red: nv4's parameter, which only its own function
writes, stops as not yet.
