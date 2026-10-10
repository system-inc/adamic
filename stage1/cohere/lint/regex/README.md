# Go regexp to stage 1 RegExp

`table.json` is the single site table: source file and line, source-unit slug,
compile operation, Go expression, fixed pattern when statically evaluable, feature
class, JS pattern and flags, and the runtime option constructor contract.
`sites.json` is its unmodified Go AST provenance. `patterns.a` contains the 82 fixed
patterns as JS RegExp literals. The two fixed comment expressions are also exported
by name. `options.a` holds the required constructor and its native gap.

The pinned cohere revision is 0183ccf8c57a775f874ca73be1adfc06e9ad1562. It has 30
non-test files importing regexp, 29 containing compile calls, 88 MustCompile sites
and 1 Compile site. Six MustCompile sites are constructed expressions, so 82
sites have statically evaluated sources and 7 are dynamic. A rule option's pattern
is now compiled as JavaScript reads it, through cohere's own `esregexp`, so those
sites are outside this Go regexp census.
Helpers and tooling retain their source-unit slug instead of inventing a rule owner.

An arbitrary accepted option-pattern language is infinite. The table inventories
all compile sites and source expressions, including each runtime option family;
it cannot enumerate every possible user string. See [gaps.md](gaps.md) for the
measured constructor and dialect blockers and the incomplete rule migrations.

| Feature class | Table rows |
| --- | ---: |
| plain | 78 |
| `(?i)` | 2 |
| `(?s)` | 1 |
| `(?m)` | 1 |
| `\p{...}` | 0 fixed sites |
| `\z` | 0 fixed sites |
| `(?P<name>)` | 0 fixed sites |
| dynamic expressions/options | 7 |
| total | 89 |

Go's ASCII whitespace becomes explicit classes, dot excludes only LF unless
`(?s)` is active, end anchors use an absolute-end assertion, LF multiline anchors
use lookaround, and braced hex escapes become JS Unicode escapes. Case-insensitive
boundaries use case-disabled scoped assertions if present. Translation happens
only in the port-time generator, never in a rule matcher.

The comparison checks every whole match and its UTF-16 start/end, including empty
matches with Go's adjacent-empty enumeration rule. The oracle alone maps Go's byte
spans through a single UTF-16 helper. Rule finding positions are not involved yet.
Captures, named-group objects, replacement APIs and invalid UTF-8 are not covered.
No universally faithful translation of arbitrary Go options is claimed. All fixed
sites agree on the recorded corpus; that is bounded evidence, not a proof over all
Unicode versions or all input strings.

The requested quiet-hundred manifest is absent on this base. The explicit alternate
100-file corpus uses 77 TypeScript compiler sources and 23 stage1 sources, taking up
to 25 distinct ASCII identifiers and 25 comments from each plus Unicode and anchor
controls. `testdata/corpus-files.json` records paths and hashes; `corpus.json` and
`corpus.a` retain 2,104 deduplicated inputs. Comments may contain Unicode; identifier
extraction itself is ASCII. These files are not represented as the quiet hundred.

```sh
source /workspace/adamic-tools/env.sh
go run stage1/cohere/lint/regex/testdata/inventory.go cohere/internal/lint/rules > /tmp/regex-sites.json
python3 stage1/cohere/lint/regex/testdata/generate.py > /tmp/regex-generate.log 2>&1
go test ./stage1/cohere/lint/regex -count=1 -v -timeout=20m > /tmp/regex-tests.log 2>&1
go vet ./stage1/cohere/lint/regex > /tmp/regex-vet.log 2>&1
```

The corpus generator currently uses the recorded compiler checkout under
`/workspace/scratch/wave1-06-typescript-6.0.3`; tests replay checked-in data and need
no external checkout. Generation must be reviewed if source pins or paths change.

The census reviews, 715ba94f to 7945d102 and 7945d102 to 0183ccf8, are recorded in
[testdata/shapes/CENSUS.md](testdata/shapes/CENSUS.md). Normal `go test` executes
all 89 per-shape fixtures through TestShapeFixtures, including its mutant and native
transition controls. No surviving pattern text or translation was silently edited.
