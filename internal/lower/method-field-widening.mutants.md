# Mutant for a method seen as a writable field (#9dmh4em)

internal/lower/invariance.go, `widenedProperties`. Old:

```go
		if viewed.Flags&ast.SymbolFlagsMethod == 0 && inside.Flags&ast.SymbolFlagsMethod != 0 && !l.hostDeclaredSymbol(inside) && !fresh
```

New:

```go
		if false && viewed.Flags&ast.SymbolFlagsMethod == 0 && inside.Flags&ast.SymbolFlagsMethod != 0 && !l.hostDeclaredSymbol(inside) && !fresh
```

Run: `go test ./internal/lower -count=1 -run 'TestMethodSeenAsWritableFieldIsRefused' -v`

Red: TestMethodSeenAsWritableFieldIsRefused's four refused rows, .a and .ts, an object literal's method and a class's, each `got <nil>` (Codex 01a1c6b1 at 7725a0bd); the six controls pass.
