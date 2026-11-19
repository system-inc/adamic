# The IR's tags reach the runtime's key through adamic.h: a mutant (#tb722h2 H2)

The mutant is one edit, applied alone, then reverted. Run
`go test ./internal/native_runtime -run '^TestAnIRTagEditedInTheHeaderMovesTheKey$' -count=1 -v`.
Not run yet: written on a machine that doesn't build Go, so "goes red" below is what the code says, not what a run
printed.

## M1. runtimeSourceParts passes over the file named adamic.h

File: internal/native_runtime/library.go, `runtimeSourceParts`, the loop over files.

Old:

```go
	for _, file := range files {
		parts = append(parts, fmt.Sprintf("file %s %x", file.Name, sha256.Sum256(file.Contents)))
```

New:

```go
	for _, file := range files {
		if file.Name == "adamic.h" {
			continue
		}
		parts = append(parts, fmt.Sprintf("file %s %x", file.Name, sha256.Sum256(file.Contents)))
```

With the ir-tags part gone from the key (v8), adamic.h's bytes are the only way the tags reach it, so neither key
holds the edited enum line any more.

Expected: TestAnIRTagEditedInTheHeaderMovesTheKey goes red with both errors, `a tag's number edited in adamic.h left
this machine's runtime key where it was` and `a tag's number edited in adamic.h left the portable parts' key where it
was`, since runtimeKeyParts and PortableKeyParts both read their files through runtimeSourceParts.
