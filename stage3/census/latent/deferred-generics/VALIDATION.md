Built on current origin/main with compiler commits a97e7b32 and 288fd3b7 cherry-picked as 1416aee7 and b9dfcf50.
Fixed deferred eligibility/counting in latent and predicate report tools; deferred body bytes stay outside examined and blocked totals.
Added census-only concrete body-entry witnesses, never-reached coverage gaps, and function-declaration scope documentation.
Caught missing-gap, attempted-count and byte-credit mutants with report and independent byte-mask audits; controls and Go tests pass.
Not covered: a whole-compiler census, new corpus numbers, all concrete specializations, backend semantics, or the whole-repository gate.

Task: #kaxjrp9. Branch: `typescript/stage3-census-deferred-generics`.
Compiler oracle and review files are unchanged after the two requested cherry-picks.

## Scope and observations

The run owns a witness ledger outside rollback state. `Latent` buffers file
records until all callers finish. The scratch overlay observes `instantiateFunction`
only after all binders have concrete checker mappings and records body entry
through `lowerFunction`, including nested declarations. The flag is an attempt
witness, not successful compilation or complete specialization coverage.
The control puts the caller in a later file, distinguishing run-wide evidence
from per-file evidence. Generic methods, arrows and class parameters do not
trigger deferral. Diagnosed generic declarations retain their diagnostics.

Legacy reports, full reports, predicate census copies, meter/rerun audits and
speculative reports now use explicit unit eligibility. Never-reached units have
a named ledger and cannot disappear without an audit failure. Body-byte unions
remove nesting overlap. Even reached deferred bodies stay in the deferred byte
bucket because a call witness alone does not prove examination of every byte.

`tsc-closure/measure.py` retains its historical hidden-tool positional argument
for CLI compatibility and uses the local `tsc-closure/hidden.py`. The historical
`stage3/census/hidden` directory is absent both in this tree and in the pin used
by closure preparation. The new calculator retains blocked-minus-independent
coverage arithmetic and adds an exclusive deferred-body bucket. The independent
rerun audit reconstructs its own byte masks; it does not call the calculator.
Historical saved reports and corpus numbers were not rewritten.

On the generated two-file control, full and speculative ledgers have six
deferred declaration units, three reached and three never reached. Their union
is 185 bytes; nested spans are not added twice. The speculative report records
406 examined bytes out of 591 source bytes, and stock TypeScript independently
matches all 20 finding depths. These are fixture results, not compiler-corpus
measurements.

The compiler now uses `ir.SortedMap[int,bool]` for optional parameters. The
census snapshot's plain-map assertion caused runtime panics. Preserving the
field's actual type fixes the compatibility seam. Restoring the old assertion
in a successfully built scratch mutant fails the speculative continuation/depth
control. The legacy continuation control's old debugger expectation was stale;
its two existing non-null assertions supply the actual current refusals.

## Commands and results

Every test wrote to a log. All invocations used hard command limits, with
`go test -timeout 90s`. Go test package times were below 60 seconds.
The commands below were run from the repository root after sourcing
`/workspace/adamic-tools/env.sh`; Node oracle commands used
`NODE_PATH=/tmp/kaxjrp9-api/node_modules`, installed from the checked-in
`stage3/api` package and lock files in scratch.

```text
export GOPROXY='https://proxy.golang.org|direct'
timeout 600s bash cloud/setup.sh
  exit 124 during TypeScript dependency fetch/checkout; setup-initial.log.txt
git fetch --depth=1 --filter=blob:none origin d92d9bfee114c80be2c375d72edae966176e3a4f
  run in the TypeScript submodule; exact pin fetched
git sparse-checkout set --no-cone '/tsc/' '!/tsc/testdata/' '/src/lib/'
git checkout --detach d92d9bfee114c80be2c375d72edae966176e3a4f
  dependency sources available; no gitlink changes delivered
timeout 240s bash cloud/setup.sh
  exit 124 during cold Go cache warming; setup-retry.log.txt
timeout 180s bash cloud/setup.sh
  exit 0; setup-final.log.txt
  Go 0.016s, Node 0.018s, submodules 0.048s, Markdown 0.050s,
  clang 0.124s, shared cache 15.476s, build 73.637s,
  cache warm 73.709s, done 73.733s; nproc 5, CPU quota 4

python3 stage3/census/latent/make_overlay.py /workspace/adamic /tmp/kaxjrp9-overlay
gofmt -w /tmp/kaxjrp9-overlay/*.go
go build -buildvcs=false -overlay=/tmp/kaxjrp9-overlay/overlay.json -o /tmp/kaxjrp9-census ./stage3/census/latent/tool
  final generation and build pass; cold first build had hit its 180s limit

timeout 90s python3 stage3/census/latent/audit_deferred.py /tmp/kaxjrp9-census /tmp/kaxjrp9-deferred
  PASS all three modes, later-file calls, unused and diagnosed generics,
  reached nested body and nested gap, explicit counts and bytes
timeout 60s python3 stage3/census/latent/audit.py /tmp/kaxjrp9-census
  PASS continuation, body ranges, attribution and no-output controls
timeout 60s python3 stage3/census/speculative/audit.py /tmp/kaxjrp9-census /tmp/kaxjrp9-speculative-controls
  PASS depths 0/1/2, sibling depth 0, no-stubs exact full JSON

timeout 120s python3 stage3/census/latent/prove_identity.py /workspace/adamic /tmp/kaxjrp9-identity b9dfcf50
  PASS C 6907 bytes, JS 9930 bytes; base, overlay-off and flag-on production identical
timeout 120s python3 stage3/census/speculative/prove_identity.py /workspace/adamic /tmp/kaxjrp9-overlay /tmp/kaxjrp9-census /tmp/kaxjrp9-ast-identity-final
  PASS exact AST identity and token-position mutant

timeout 120s go test -count=1 -timeout 90s ./stage3/census/...
  final exit 0; all census packages pass
  the first cold invocation hit its outer 120s build limit, not a test timeout

node stage3/census/speculative/stock.cjs /tmp/kaxjrp9-deferred/source /tmp/kaxjrp9-deferred/speculative.jsonl /tmp/kaxjrp9-deferred/stock.json
python3 stage3/census/speculative/report.py /tmp/kaxjrp9-deferred/source /tmp/kaxjrp9-deferred/speculative.jsonl /tmp/kaxjrp9-deferred/stock.json /tmp/kaxjrp9-deferred/report
python3 stage3/census/speculative/verify.py /tmp/kaxjrp9-deferred/speculative.jsonl /tmp/kaxjrp9-deferred/report/RESULT.json /tmp/kaxjrp9-deferred/source
  PASS stock ancestry, denominator, gap inventory and all six artifact mutants

node stage3/census/latent/stock_units.cjs /tmp/kaxjrp9-deferred/source /tmp/kaxjrp9-deferred/manifest.json /tmp/kaxjrp9-deferred/stock-units.json
  PASS stock TypeScript 6.0.3, 2 files, 7 function declarations
Python fixture control calls full_report.audit_units, audit_counts and rerun2.summarize
  PASS spans, diagnostic ownership, explicit deferred eligibility, counts and gap audit
Python scratch mutant restores the old OptionalParameters plain-map assertion
  PASS mutant builds and is caught by speculative continuation/depth control

git diff --check
  PASS
```

The last two Python controls are recorded in stock-full-report.log.txt and
map-mutant.log.txt. Individual scratch logs remain under `/tmp/kaxjrp9-*`.
Lane checks and final branch delivery are reported with the delivered commit.

## Mutants and catchers

| Mutant | Catcher |
| --- | --- |
| Drop never-reached list, legacy/full/speculative | Independent gap ledger audit in audit_deferred.py |
| Count a deferred function as attempted, all three modes | Independent status-membership count in audit_deferred.py |
| Credit deferred bytes as examined, legacy/full | Independent byte-mask oracle in closure rerun audit, despite balanced aggregate subtraction |
| Drop full-report never-reached list | full_report.audit_counts deferred coverage assertion |
| Speculative drop-list or deferred-as-examined | speculative/verify.py gap ledger and independent byte coverage |
| Speculative headline depth, top-20 count, omitted source or false byte share | speculative/verify.py exact recount and inventory |
| Plain-map optional-parameter snapshot | Speculative depth/continuation audit after successful mutant build |
| First-error/no-stubs speculative continuation | speculative/audit.py nested depth and count assertions |
| Zero all failure depths | speculative/audit.py exact depth control |
| Token-position AST attribution | speculative/prove_identity.py async-root depth identity |
| Leak a stub into C or JS output | Overlay-off byte comparison, both backends |
| Signature included in body eligibility | latent/audit.py signature/body range assertion |
| Extra NotYet attributed to the wrong file | latent/audit.py sibling-file equality |

The planted overlay-only NotYet is observed as exactly +1 at the selected target;
the imported sibling and other file remain unchanged. No upstream corpus runs,
semantic backend runs, all-specialization proof, production compiler edits,
or updated readiness totals are claimed.
