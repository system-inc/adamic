# Go regexp to stage 1 RegExp

`table.json` is the original 107-site translation table: source file and line, source-unit slug,
compile operation, Go expression, fixed pattern when statically evaluable, feature
class, JS pattern and flags, and the runtime option constructor contract.
`sites.json` is the live Go AST census at the current pin; its delta from that
translation baseline is recorded below. `patterns.a` contains the original 83 fixed
patterns as JS RegExp literals. The two fixed comment expressions are also exported
by name. `options.a` holds the required constructor and its native gap.

The original translation baseline used cohere 7945d102a6c18dd36adf9114a758ce646e8b2359. It had 39
non-test files importing regexp, 38 containing compile calls, 89 MustCompile sites
and 18 Compile sites. Six MustCompile sites are constructed expressions, so 83
sites have statically evaluated sources and 24 are dynamic. The stated 92-site
census does not match this pin. No sites were invented or dropped to reach 92.
Helpers and tooling retain their source-unit slug instead of inventing a rule owner.

An arbitrary accepted option-pattern language is infinite. The table inventories
all original compile sites and source expressions, including each runtime option family;
it cannot enumerate every possible user string. See [gaps.md](gaps.md) for the
measured constructor and dialect blockers and the incomplete rule migrations.

| Feature class | Table rows |
| --- | ---: |
| plain | 77 |
| `(?i)` | 4 |
| `(?s)` | 1 |
| `(?m)` | 1 |
| `\p{...}` | 0 fixed sites |
| `\z` | 0 fixed sites |
| `(?P<name>)` | 0 fixed sites |
| dynamic expressions/options | 24 |
| total | 107 |

Go's ASCII whitespace becomes explicit classes, dot excludes only LF unless
`(?s)` is active, end anchors use an absolute-end assertion, LF multiline anchors
use lookaround, and braced hex escapes become JS Unicode escapes. Case-insensitive
boundaries use case-disabled scoped assertions if present. Translation happens
only in the port-time generator, never in a rule matcher.

The comparison checks every whole match and its UTF-16 start/end, including empty
matches with Go's adjacent-empty enumeration rule. The oracle alone maps Go's byte
spans through a single UTF-16 helper. Rule finding positions are not involved yet.
Captures, named-group objects, replacement APIs and invalid UTF-8 are not covered.
No universally faithful translation of arbitrary Go options is claimed. All original fixed
sites agree on the recorded corpus; that is bounded evidence, not a proof over all
Unicode versions or all input strings.

The current cohere pin is f5d1934a2d7bebe706210cb1cfd01aebff4f8ca7.
Its live `sites.json` has 89 RE2 compile sites (88 MustCompile, one Compile),
82 fixed and seven dynamic. Commit c2e39b75 moved the user-option sites to
JavaScript regexp; 0b892cc9 added the fixed class-matcher identifier pattern.
The 107-row translation table, 83 fixed port literals and shape fixtures above
remain the original port baseline in this pin-only unit; port behavior is not
changed. They do not claim complete coverage of the new live inventory.

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

The 715ba94f to 7945d102 census review is recorded in
[testdata/shapes/CENSUS.md](testdata/shapes/CENSUS.md). Normal `go test` now executes
all 107 per-shape fixtures through TestShapeFixtures, including its mutant and native
transition controls. The new pin changes one standard-library regexp site in each
direction and moves 66 source lines; no old pattern text was silently edited.
