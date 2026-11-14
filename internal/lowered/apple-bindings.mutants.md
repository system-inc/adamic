# Apple binding contents in the lowering key (slice-23)

On the Linux box I ran `TestAppleBindingsContentsAtTheSamePathMoveTheKey` with a temporary Go overlay. Removing the file loop compiled and made the test red; dropping the overlay restored the implementation. The source file was not edited by the mutant run.

| Mutant | Edit | Red |
|---|---|---|
| Binding files omitted | `internal/lowered/lowered.go`, `inputsOf`: remove `for _, relative := range bindings { ... }`, including its file read and `apple-binding` flag append; keep the directory walk and sort. | `TestAppleBindingsContentsAtTheSamePathMoveTheKey`, `internal/lowered/lowered_test.go:568`: `different binding contents at the same path kept the lowering key` |
