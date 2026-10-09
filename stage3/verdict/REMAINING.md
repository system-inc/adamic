Built: faithful temporary-filesystem execution and library-summary formatting for six exclusion groups.
Base: 34420929 on codex/stage3-verdict-harness; this unit changes only stage3/verdict/.
Observed: 220 more inputs, 240 more configurations; A 240/240 plus one authored diagnostic fixture.
Mutants: six B wrappers each fail stdout only; raw and projected output each differ by exactly one byte.
Not covered: 614 input files requiring API diagnostic collection, compiler-version overrides, canonical aliases or suggestions.

The census now selects **11,830 of 12,444 inputs**, with **13,693 configurations**.
Wholly excluded inputs fell from **834 to 614**. Every previous 13,453 selected
configuration remains selected. Independently regenerated censuses from pristine
and adapted pinned checkouts are byte-identical. Selection uses source metadata,
reference diagnostic scope and the pinned option/parser API, never the tested
binary's results. No expected baseline or driver golden changed.

The starting exclusions were:

| Reason | Inputs |
|---|---:|
| API compiler version override @typescriptversion: 5.0 | 3 |
| API compiler version override @typescriptversion: 5.5 | 1 |
| Windows virtual paths require drive and separator semantics unavailable on Linux | 23 |
| absolute package/config paths require mounted virtual roots; source text is preserved | 14 |
| absolute source references require mounted virtual roots; source text is preserved | 172 |
| all configurations need API diagnostic collection or harness formatting | 603 |
| case-insensitive virtual filesystem differs from Linux | 4 |
| harness directive @capturesuggestions: true | 1 |
| harness directive @suppressoutputpathcheck: true | 13 |

New groups were measured with the existing adapted Node 6.0.3 wrapper:

| Group | Added inputs | Added configurations | A passes | B passes |
|---|---:|---:|---:|---:|
| absolute-source | 163 | 183 | 183/183 | 0/1 |
| drive-paths | 22 | 22 | 22/22 | 0/1 |
| absolute-package | 14 | 14 | 14/14 | 0/1 |
| output-path | 13 | 13 | 13/13 | 0/1 |
| case-host | 4 | 4 | 4/4 | 0/1 |
| library-placeholders | 4 | 4 | 4/4 | 0/1 |

The case-host group has three clean cases and one option-deprecation baseline.
Its B proof uses an additional authored checker diagnostic fixture, `fixtures/case-host/input.a`: A 1/1,
B 0/1. Every other group uses a diagnostic-bearing pinned upstream case as its
fixture. `evidence/remaining/fixtures.json` records each exact source, baseline,
configuration and hash. The one-byte mutants preserve length, stderr and exit;
`prove_groups.py` checks both raw stdout against a same-directory Node control
and the projected diagnostic bytes. Each mutant is caught by stdout alone.

Absolute source references (163 inputs, 183 configurations) retain imports,
references and JSX pragmas unchanged. Linux bubblewrap runs the actual command
line in a mount namespace whose writable case roots come from the temporary
output directory. Pinned `tests/lib` is mounted at `/.lib`; unmodified generated
TypeScript declaration inputs are mounted at `/.ts`. Runtime installation paths
remain available read-only. No host-root file is created or overwritten. Nine
of the original 172 absolute-source inputs expose additional API diagnostic
collection requirements once their paths can be resolved; those remain excluded.

Drive paths (22 inputs) preserve drive-prefixed compiler arguments. TypeScript's
own path parser supplies drive geometry; on Linux the physical paths are colon
named directories under the virtual working directory. Metadata backslashes
become forward slashes, as in the upstream harness. Program text stays unchanged.
One input, `commonSourceDir3.ts`, mixes A:/ and a:/ under a case-insensitive host;
the real CLI's canonical-name policy cannot be changed through a flag, so it
remains excluded rather than treating two directory aliases as equivalent.

Absolute package/config paths (14 inputs) use the same namespace, including
absolute symlink targets so realpath identity matches the virtual filesystem.
For projects at filesystem root, the runtime mounts must not participate in
CLI default root globs. The pinned API's exact input-file list is appended to
that temporary JSONC project's `files` property. Original option text and all
original diagnostic byte positions remain intact, including trailing comments.
A lexical scanner finds the actual object end rather than braces in comments.
Emit destinations under /bin or /lib are writable temporary paths; dynamic
loader subtrees remain read-only. Both root-glob and symlink cases pass A.

Output-path overrides (13 inputs) suppress the API-only overwrite check by
sending CLI emit to a private `.verdict-emit` directory. Original checker and
emit flags remain active. These cases have no explicit output destination;
combining this override with explicit outDir/outFile is rejected. This suite
judges diagnostics and exit, not the emitted-file baseline.

Case-host directives (four inputs) are admitted when input path prefixes have
no case-folding aliases. At this pin their diagnostic behavior does not depend
on a case-insensitive filesystem. The conflicting drive case is explicitly
rejected; its canonical alias mutant exercises that guard. Emitted source-map
metadata and platform-dependent casing metadata are outside this diagnostic suite.

Library placeholders (four inputs) reproduce the upstream formatter's
`lib.*.d.ts(--,--)` headers and virtual /.src-before-/.ts diagnostic ordering.
The formatter uses CLI library diagnostics from the tested binary and retains
all diagnostic message bytes, codes and source positions. It never supplies
expected messages. The B wrapper's uppercase Error byte still survives the
formatter and fails comparison. Independent code, filename and message-byte
mutants also fail the formatter fixture. Raw output is always retained.

The remaining exclusions are disjoint:

| API requirement | Inputs |
|---|---:|
| parser errors suppress other diagnostics on CLI; API collects both | 327 |
| global/options diagnostics suppress semantic diagnostics on CLI; API collects both | 189 |
| JavaScript syntax errors suppress semantic diagnostics on CLI; API collects both | 50 |
| global/options diagnostics suppress semantic diagnostics on CLI; API collects both; parser errors suppress other diagnostics on CLI; API collects both | 22 |
| config option diagnostics suppress semantic diagnostics on CLI; API collects both | 10 |
| pre/post emit consistency diagnostics exist only in the API harness | 5 |
| config parse diagnostics are discarded by API compiler runner but reported by CLI | 2 |
| option diagnostics suppress global diagnostics on CLI; API collects both | 1 |
| parser errors suppress other diagnostics on CLI; API collects both; pre/post emit consistency diagnostics exist only in the API harness | 1 |
| global/options diagnostics suppress semantic diagnostics on CLI; API collects both; option diagnostics suppress global diagnostics on CLI; API collects both | 1 |
| Version override: @typescriptVersion 5.0 | 3 |
| Version override: @typescriptVersion 5.5 | 1 |
| Case-insensitive canonical drive aliases | 1 |
| @captureSuggestions | 1 |
| **Total** | **614** |

The first ten rows total 608 API-diagnostic inputs. Parser, JavaScript syntax,
global, option and config errors make the CLI stop before diagnostics that the
upstream API harness explicitly collects together. Re-running a modified input
or deleting options would change the tested program or semantics. Config-parse
errors are intentionally discarded by the API runner but printed by the CLI.
Pre/post emit consistency diagnostics are assertions implemented by the
in-process harness. Historical-version overrides mutate the compiler's API
version state, not a command-line setting. Suggestions are an API diagnostic
category absent from tsc CLI output. The canonical alias case requires the
compiler host's getCanonicalFileName policy, not just filesystem symlinks.
These require an in-process adapter for identical upstream errors baselines.
Individual filenames and known excluded configurations remain in selection.json;
`evidence/remaining/counts.json` lists filenames in each disjoint API subgroup.
This is a scope inference from pinned harness/CLI control flow, not an observed
native pass rate. Native and typescript-go were not run in this unit.

Commands run (all test output went to files):

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/verdict-remaining-setup.log 2>&1
source /workspace/adamic-tools/env.sh
# In a pristine pinned external TypeScript checkout:
npm ci --ignore-scripts --no-audit --no-fund > /tmp/verdict-remaining-upstream-npm.log 2>&1
npx hereby local > /tmp/verdict-remaining-upstream-build.log 2>&1
export STAGE3_VERDICT_NODE_TSC=/tmp/stage3-verdict-adapted/built/local/tsc.js
PYTHONDONTWRITEBYTECODE=1 python3 stage3/verdict/census.py   /tmp/stage3-verdict-auto-upstream/upstream   /home/agent/.cache/adamic-stage3/api/node_modules/typescript/lib/typescript.js   /tmp/verdict-remaining-final.json > /tmp/verdict-remaining-census.log 2>&1
PYTHONDONTWRITEBYTECODE=1 python3 stage3/verdict/census.py   /tmp/stage3-verdict-adapted   /home/agent/.cache/adamic-stage3/api/node_modules/typescript/lib/typescript.js   /tmp/verdict-remaining-independent.json > /tmp/verdict-remaining-independent.log 2>&1
# Snapshot the prior census before the unit changes it:
git show 34420929:stage3/verdict/selection.json > /tmp/verdict-834-before.json
PYTHONDONTWRITEBYTECODE=1 python3 stage3/verdict/prove_groups.py   /tmp/verdict-834-before.json stage3/verdict/selection.json   /tmp/stage3-verdict-auto-upstream/upstream /tmp/new-remaining-proof   > /tmp/new-remaining-proof.log 2>&1
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s stage3/verdict   -p 'test_*.py' > /tmp/verdict-remaining-tests.log 2>&1
```

The executed final proof used `/tmp/verdict-remaining-v5.json` (byte-identical to
the committed selection) and `/tmp/verdict-remaining-final-proof` as output.
The case-host proof was then rerun with `--group case-host` against the
committed selection into `/tmp/verdict-remaining-case-host-fixture`, explicitly
using the authored checker-error fixture instead of the existing option error.
Its A 4/4 plus fixture A 1/1 and B 0/1 are the case-host rows in the consolidated
proof. `checks-final.log.gz` records the final focused test run with
`STAGE3_VERDICT_UPSTREAM` set to the pristine checkout.
All commands exited 0; focused tests: **23 passed, no skips**. Resource mutants
append a byte separately to copied pinned tests/lib and generated declaration
inputs; each fails its corresponding SHA256 preflight before compiler execution.
The JSONC fixture catches a last-brace-in-comment mutant; path and library
fixtures exercise canonical aliases and formatter mutations. No whole-package
tests or full gate were run. No Adamic oracle fixtures were added, so the
repository oracle counts table is unchanged.

Setup reported cumulative timings: Node 0.033s, Go 0.034s, submodules 0.080s,
markdown 0.090s (dependency step 0.009s), clang 0.203s, Go build 36.061s,
tests deferred 36.169s, build cache 36.170s, total 36.196s. `nproc`: **5**,
cgroup quota: four CPUs. Full logs, group reports, fixture captures and resource
build provenance are in `evidence/remaining/`. Earlier failed probes exposed
root glob leakage, read-only emit roots, symlink identity and JSONC trailing
comment handling; the final group proof was rerun after fixing those issues.
Previous unit measurements remain historical in REPORT.md, UPSTREAM.md and
TYPESCRIPT_GO.md; their pass totals describe their earlier censuses.
