# Mutant for a method seen as a writable field (#9dmh4em)

internal/lower/invariance.go, `widenedProperties`. Old:

```go
		if viewed.Flags&ast.SymbolFlagsMethod == 0 && inside.Flags&ast.SymbolFlagsMethod != 0 && !fresh
```

New:

```go
		if false && viewed.Flags&ast.SymbolFlagsMethod == 0 && inside.Flags&ast.SymbolFlagsMethod != 0 && !fresh
```

Run: `go test ./internal/lower -count=1 -run 'TestMethodSeenAsWritableFieldIsRefused' -v`

Red: MEASURED
