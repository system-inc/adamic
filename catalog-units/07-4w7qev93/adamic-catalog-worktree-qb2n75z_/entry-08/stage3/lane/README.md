# Stage 3 landing lane

Run from the checkout that integration intends to land, with Node 24.19.0,
Git, npm and Python 3 available:

```sh
stage3/lane/run.sh /absolute/path/to/new-results > /absolute/path/to/lane.log 2>&1
```

The results argument is optional. Without it, the lane creates a fresh temporary
directory and prints its absolute path on the first line. With an argument,
the results directory must be new. The verdict remains the last stdout line.
The lane calls this checkout's unmodified
`stage3/apply.sh` with `results/adapted-tree`, then its unmodified
`stage3/oracle/run.sh` with that tree and `results/oracle`. Apply's normal
`STAGE3_CACHE` override also selects the pinned TypeScript compiler API used
by the declaration guard. All command output goes to logs. The lane writes
`execution.json`, `report.json`, a copy of the adapted tree's `patch-set.md`,
and a one-line `verdict.txt`; its own exit is
zero only for PASS. Existing results are never replaced. To replay the guard
against preserved results, run `python3 stage3/lane/check.py RESULTS`.

`expected.json` deliberately permits exactly one upstream failure:
`unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`.
Both the reporter title and `baseline_diffs == ["api/typescript.d.ts"]` must match.
Apply, install and build must exit zero; upstream tests and oracle must exit
one. All runners, eight workers and no test filter are
required. A timeout,
another single failure, missing evidence, or an unexpected exit fails the lane.
Counts are checked against both oracle JSON and the final reporter summary.
Reporter titles retain their exact words and hierarchy, with the indented
reporter lines joined by one space and the trailing display colon removed.

Platform keys use Node's `process.platform`, `process.arch`, and `process.version`.
Unknown combinations fail before apply. Linux x64 with Node v24.19.0 requires
106,366 passing, one failing and zero pending. The darwin arm64 entry requires
106,369 passing, one failing and zero pending, from @system_adamic's measurement
on area c193801. This worker has measured Linux only. Expectations are checked
in so count changes require an intentional edit.

## Generating the sanction

The manifest is generated from the pinned pristine snapshot, independently of
any full-apply oracle diff. After apply has installed the stock API dependency:

```sh
NODE_PATH="${STAGE3_CACHE:-$HOME/.cache/adamic-stage3}/api/node_modules" \
  python3 stage3/lane/generate-api.py \
  "${STAGE3_CACHE:-$HOME/.cache/adamic-stage3}/typescript.git" \
  > /tmp/lane-generate.log 2>&1
```

The argument can also be a pinned TypeScript checkout. Generation reads pinned
Git objects, rather than trusting modified checkout files. The reconstruction
uses these adaptations' mechanical proof data:

- 20: the 189 optional owner paths in 70's recorded composed API attribution
  (`evidence/wave6/api-attribution.json`), proved by the same parsed-owner
  reconstruction as `20-optional-declarations/check-baselines.cjs`. Add undefined
  to each exact original AST type, retaining the function/conditional parentheses.
- 30: `remaining-api.json` proves only the existing 189/28/1 owners and
  contributes no additional public owner. Its README hands the optional types
  in types.ts to 32 rather than approving a consumer rewrite.
- 32: `handoff-sites.json` and `owner-handoffs/api.json` identify the public
  `AmdDependency.name` and `CommentRange.hasTrailingNewLine` property unions.
  `public-host-sites.json` and `resumed-watch/api.json` additionally prove
  `BuilderProgramHost.createHash`, which was already a property. Match exact
  owner paths, original AST types and before/after proof data. Include only
  existing optional properties whose change is exactly an undefined union;
  preserve function-type parentheses as adaptation 20 does. Original methods
  are excluded even when the host ledger labels them as public owners.
- 33: read every `api-projection.json` under its proof directory. Each records
  zero additional public changes; its AllDecorators and RawSourceMap handoffs
  are internal. A newly recorded public change requires explicit reconstruction.
- 40: `class-rules.json`, with the composed proof's 27 public brands and the
  `ErrorCallback.arg0` payload. Resolve source owners and original any tokens
  before replacing their emitted types, exactly as adaptation 30's composed
  `check-api-baselines.cjs` does. This is 28 distinct API lines.
- 70: its recorded `readonly_owners` proof, exactly `setTextRange.location`.
  Wrap the existing TextRange reference in Readonly, preserving undefined.
- 75: its wave-2 `api.json` public addition. Its own `api-additions.cjs` must
  remove that exact optional property and recover the prerequisite byte for byte.

Generation asserts the 189/0/3/0/28/1/1 counts for 20/30/32/33/40/70/75, unique owner lines and original AST
forms, and records the proof files' SHA-256 hashes. `sanctioned-api.json` includes
the exact generated diff and normalized before/after declarations. The raw
manifest diff is provenance; runtime approval uses normalized declarations.
No API lines are selected from the observed area diff or entered by hand.

## Declaration normalization

First require `oracle/baseline.diff` to exactly describe the reference and local
API files, including its file headers, hunks and context. Extra diff files,
missing lines, or tampered context cannot bypass the declaration guard.

Parse the pristine, reference and local snapshots with stock TypeScript 6.0.3.
Record each namespace/interface header and member under its owner path;
record type aliases and other declarations whole. Number overloads in order.
Each normalized line retains AST kinds, identifiers, literal values, modifiers,
question tokens, signatures, heritage, type arguments and ordered union members.
Ignore source positions, whitespace, quote delimiters and redundant type
parentheses. The leading `|` of a multiline union is only a parser separator;
`a | { ... }["trace"]` and the same union spread over lines normalize equally.
Comment wording is retained as a separate normalized entry, allowing whitespace
reflow but rejecting an extra comment or changed documentation.

Sort the owner keys and compare their original/composed normalized lines exactly
with the generated sanction. An extra owner, missing owner, changed type or
changed declaration kind fails with its owner named. A method-to-property
conversion retains its changed AST kind; formatting normalization does not
approve that structural change.

Some adapters already incorporate proved changes into their reference snapshot;
current apply incorporates 40's 28 changes. Those reference changes must be an
exact subset of the manifest. The complete pristine-to-local composition must
still contain every sanctioned change and no others. This checks the union
without requiring already-accepted reference lines to reappear in the diff.

## Tests and integration boundary

```sh
python3 -m unittest discover -s stage3/lane -p 'test_*.py' -v > /tmp/lane-tests.log 2>&1
bash -n stage3/lane/run.sh
```

Tests plant valid API diffs with an unsanctioned member and a missing sanctioned
member, an incorrect passing count, and a different single failure title.
Additional tests verify the three property handoffs and exclusion of method conversions,
and cover changed types, a multiline indexed-access union, accepted
and unproved reference changes, comment additions, source pin, diff tampering,
platform/version mismatch, missing evidence, exits, pending counts and filtered
runs. They execute the same checker CLI on isolated declaration snapshots and
use the real upstream reporter's failure title and count output.

Integration owns wiring the lane into every area/stage3 merge and main push
that touches stage3. This unit changes no adaptation, compiler, shared oracle,
workflow, or fixture bucket. Ordinary `apply.sh` writes its generated table
inside the adapted output tree; only explicit `--write-table` also updates
the checked-in `stage3/patch-set.md`. The lane uses ordinary apply and
preserves a copy beside its report.json.
See REPORT.md and evidence for the measured area-tip verdict.
