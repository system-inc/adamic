Built: no new rule implementation; the configured checks remain at f114b002.
Commits: prior pushed tip f8e7f2bb; this update records fresh origin and regexp evidence.
Commands: fetch all origin heads succeeded; independent Go and Node probes exited 0 with empty stderr.
Mutants: no production change or new mutant; prior configured/default rule and lifetime mutants remain recorded in README.md.
Blocked: general Go-compatible regexp matching remains unfinished; no additional rules claimed.

Origin was refreshed on October 7, 2026. Main is now
`ef3d907ecdc4c771b016f7d9c52372def057a340`, the bridge branch is
`5afbdb83da2ed7ad9815657cd3f6ececd5294bf6`, and the newly available harness
branch is `2650ad595b82220c368631ea13139fad4b306ed6`.

The harness branch accepts `.a` rule modules and adds emitted-JavaScript
comparisons to the syntax-only registry harness. This supersedes the earlier
observation that it was absent from origin. It does not supply a JavaScript
adapter for this type-aware C checker library. Main's CLI still explicitly says
`tsgo is an external native checker library; JavaScript is not supported`.
Main now contains `internal/native/regexp.go` and its runtime, superseding the
constructor-support gap for that compiler. This worker branch remains on its
original required bridge base; no compiler or harness integration was performed
and no shared files were edited.

Replacing the explicit missing matcher with ECMAScript `RegExp` would still
violate byte agreement with Go cohere. Its production id-match decoder compiles
patterns with Go's regexp package. The independent Go and Node observations are:

| Pattern | Input | Go regexp | Node RegExp with `u` |
| --- | --- | --- | --- |
| `\Afoo\z` | `foo` | matches | invalid pattern |
| `^foo$` | `foo` followed by LF | no match | no match |
| `.` | carriage return | matches | no match |
| `(?i)^foo$` | `FOO` | matches | invalid pattern |
| `^[^_]+$` | `foo` | matches | matches |

The last row controls the exact test-only matcher used in the earlier 246-input
comparison. That measured scope remains valid; it is not a general pattern port.
Both new probe processes exited 0 with empty stderr. Their complete streams are
[Go](validation/regexp-go.stdout) and [Node](validation/regexp-node.stdout).
The Go probe source is [regexp_semantics.go](testdata/regexp_semantics.go).

```sh
source /workspace/adamic-tools/env.sh
git fetch origin --no-recurse-submodules \
  '+refs/heads/*:refs/remotes/origin/*' > /tmp/wave29-next-fetch.log 2>&1
go run stage1/cohere/typeaware/wave-29-configured/testdata/regexp_semantics.go \
  > /tmp/wave29-regexp-go.stdout 2> /tmp/wave29-regexp-go.stderr
```

Node evaluated the same five pattern/input pairs with
`new RegExp(pattern, 'u').test(text)`, catching constructor errors as `invalid`.
Neither tool imports Adamic or the bridge. This probe demonstrates semantic
incompatibility; it is not a new production check or a mutation kill.

All origin typeaware claim Markdown blobs were inspected: 33 distinct blobs
mention 114 entries in the 197-rule VOLUME population. There are still entries
not claimed, so this update does not claim that the ranking is exhausted. The
existing claim's general id-match execution remains incomplete, and the earlier
instruction says to stop on other blockers instead of editing shared files.
No next rules were reserved. No additional corpus/performance run was needed
for this probe and documentation update; the passing comparisons, sanitizer
checks, mutants and timings in README.md describe the unchanged rule code.
