# Local gates for the pushed candidate

Compiler commit `154e142d26a991681907c2a8c2b649795f913b7f`. The exact SHA was published on request while the full lower rerun was still in progress so Loom could start in parallel.

| Command | Result |
| --- | --- |
| `gofmt -l cmd internal` | Exit 0, no files |
| `go vet ./...` | Exit 0 |
| `go test -count=1 -timeout=30m ./cmd/adamic-gate/...` | Exit 0, both packages |
| `go test -count=1 -timeout=90m -v ./internal/lower/` | Exit 1 after 3437.974s, with the pinned TypeScript checkout in `ADAMIC_TYPESCRIPT_SOURCE` |
| `go test ./internal/lower -run '^TestLiteralReceiverFreshness' -count=1 -v` | Exit 0 on the identical lowerer before the JavaScript-only rebase |

The full lower rerun had one failing top-level test, `TestOptionalMemberPresenceLowers`. The guarded receiver case fails in both extensions when its interface uses the writable optional function field `m?: () => number`. In this case `m() { return this.n; }` reads a readonly number and does not expose its receiver. The classifier nevertheless treats the literal as held, and the existing check refuses to turn its required method into a writable field that can become undefined. The method-signature form of the same test passes. A focused census of the failing source accepted both extensions on the baseline and reproduced the candidate's refusal in both formats.

The baseline fixture runs used the oracle helpers through a temporary test. The candidate adds no oracle fixtures or IR state.
