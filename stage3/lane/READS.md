# Lane product and test reads

Planner inventory for Task: #hefhych. All paths are repository relative. `**`
includes dotfiles and every descendant. This records Loom unit reads, which
must be keyed even when buildcache already keys the fetched product. Go sources
and their import closure are supplied by Loom. Do not infer that a file omitted
from buildcache.Inputs is safe to omit from a shard's Loom unit key.

## Products and their key tests

Install means TestProduct_Stage3LaneInstall; Compiler and Harness are the other
two top-level TestProduct_ declarations. Their TracksInputs tests copy and key
the same declared input sets in private directories. Fetch in the shard computes
the same Harness inputs but never invokes a product worker.

| Path or glob | Readers | Observation and buildcache reconciliation |
|---|---|---|
| `stage3/**`, excluding `stage3/lane/**`, `stage3/oracle/**`, every `__pycache__` subtree, `stage3/api/node_modules/**` and `stage3/typescript/node_modules/**` | All three products and TracksInputs tests; shard via artifact.input_key and final fetch | Empirical product/shard trace plus source audit. This is the intentionally broad existing artifact identity. Includes all adaptation CJS, JSON, .a proofs, compressed evidence, reports, patch tables, apply products' READS.md/evidence, `stage3/apply.py`, `stage3/caches.py`, `stage3/source.json`, both NodeProjects' package.json/package-lock.json, `stage3/api/node_modules.sha256`, and all other non-excluded stage3 files. laneInputs walks this exact set; artifact.input_key reads its bytes twice per worker and once in shard verification. Future files in this set are inputs too. |
| `stage3/lane/products.py`, `stage3/lane/artifact.py`, `stage3/lane/node_modules.sha256` | All three products and TracksInputs tests; shard's final-product fetch; artifact.py and pin also read by shard verification | Explicit Inputs.Files. Worker Python imports and source/hash/pin reads observed. The pin is also an Inputs flag and upstream_packages field of the artifact identity. |
| `stage3/typescript/node_modules/**` | Install worker payload reads; all artifact.input_key calls may enumerate directories | installed_packages opens the hidden .package-lock.json, the tree package-lock.json and each installed package.json; packages_digest opens every regular file and records all directory names, modes and symlinks. Then copytree reads the payload. Physical install excluded from Inputs.Files and artifact.input_key: the declared lane/node_modules.sha256 verifies every byte before copying, and tree package-lock.json is checked against both installed inventory and the adapted predecessor lock. Empirical install trace covers the physical reads. |
| `stage3/api/node_modules/**` | Harness worker payload reads, plus apply predecessors; all input_key calls may enumerate directories | pinned_packages validates all bytes and copytree bundles stock TypeScript under api/. Represented by the declared api/node_modules.sha256 and API lock, not physical installation or runner marker. Source audit supplements the install trace. |
| `go.mod` | All product/key operations | Observed root-discovery open; part of Go closure. Absent stage3/go.mod and stage3/lane/go.mod probes do not supply bytes. |

The broad set deliberately includes files that a compiler subprocess does not
execute: artifact.input_key hashes their contents. It is not correct to reduce
that read line to adapt.cjs alone. TracksInputs copies include non-Go files in
that set even if Git's clean-index optimization avoids reopening them while
computing the original key.

## Shard 001

The shard reads every row needed by final-product fetch and artifact identity
above, except the physical NodeProject installs. It additionally reads:

| Path or glob | Reader and purpose | Observation / product-input reconciliation |
|---|---|---|
| `stage3/lane/shards.json` | TestShard_Stage3LaneFile001 reads command argv and expect counts | Observed Go open. Unit-only input: no product worker reads the manifest. Missing and empty manifest negatives use t.TempDir files. |
| `stage3/lane/shard-plan.json` | shard.py loads tasks, names digest, counts, baseline expectations and inputs hash | Observed Python open; unit-only input, intentionally not a product input. |
| `stage3/lane/shard.py`, `stage3/lane/shard_plan.py`, `stage3/lane/view.py`, `stage3/lane/check.py` | Imported/executed shard runtime and runtime_hash | Observed source opens. Unit-only inputs except artifact.py already listed. |
| `stage3/lane/run.py`, `stage3/lane/merge_shards.py`, `stage3/lane/prepare_shards.py`, `stage3/lane/normalize-api.cjs`, `stage3/lane/expected.json`, `stage3/lane/sanctioned-api.json` | shard_plan.RUNTIME hashes every file; expected.json also checked after running tests | Observed opens even where a conditional API normalization path does not execute for 001. Unit-only inputs. |
| `stage3/oracle/run.py`, `stage3/oracle/profile-tasks.cjs`, `stage3/oracle/run-prepared.mjs`, `stage3/oracle/run-tasks.cjs`, `stage3/oracle/observe-tests.cjs`, `stage3/oracle/observe-errors.cjs`, `stage3/oracle/failure_details.py` | runtime_hash; run-tasks.cjs and observation modules execute; failure_details imports | Observed source/hash opens. Oracle subtree is excluded from product identity but mandatory in the shard unit key. |

This is the complete shard_plan.RUNTIME list, not just executed imports. The
Go manifest-negative test reads only its freshly created private manifest and
Go closure. No other top-level Go shard is declared. The manually run Python
fixture/proof suite also reads its own `stage3/lane/test_*.py`, `prove_*.py`,
`bind_products.py`, manifests and `stage3/lane/evidence/**`; it is a verification
command rather than another declared Go unit. If Planner makes that suite its
own unit, use `stage3/lane/**` conservatively, plus the broad stage3 and oracle
sets above. The fixture source audit is supplementary to the product/shard trace.

## Trace evidence and reconciliation

`evidence/planner-reads/install-paths.tsv.gz` and `shard-paths.tsv.gz` list every
repository path opened or probed, not just the paths expected from source.
Their complete traces, stdout logs and both extraction parsers
are retained. PLANNER_FIXES_REPORT.md records exact trace commands, exits and
separate untraced cold measurements. Trace commands execute compiled Go test
binaries, separating runtime access from Go compiler access.

Repository `.git/**`, `.gitmodules`, `cohere/.git`, `cohere/.gitmodules` and
`cohere/TypeScript/.git` reads in the lists come from buildcache's generic Git
index/status machinery, including submodule bookkeeping and unsuccessful
config/sequencer probes. These are not undeclared compiler sources. Go module
metadata belongs to the Go closure. The lists retain all such paths for review;
Planner should not key mutable Git indexes, object-pack filenames or logs as
lane semantic inputs. Absent local npm config and Python bytecode probes supply
no bytes. Python writes no bytecode. A future checkout-local npm config must be
refused or declared rather than silently accepted.

Products read predecessor directories and the shard reads fetched ready.json,
shards-ready.json, api/** and tree/**, including the complete payload verification
and then private view baselines/built harness/package payloads. These live outside
the source repository and are product dependencies, not additional repository
read globs. Shard outputs and observation logs are private t.TempDir data.
Tool executables, Node/npm/Python libraries and OS files are outside the repository;
version reports participate in Inputs.Toolchain. Product npm runs offline with
private user/global config and cache, and never writes HOME.

Observed-only versus declared differences are accounted for as follows:
physical installs are strictly validated against declared byte pins; generated
product/view data is covered by predecessor identity; Git/module infrastructure
is intrinsic key machinery; shard manifests and all runtime files are additional
Loom unit reads, not product-worker inputs. Declared-only broad stage3 files and
Go recipe source are intentional conservative keying. The apply product-99
metadata omission found during this audit was fixed in the parent branch.
Planner still needs to generate and prove its unit read lines; these documents
do not claim that Loom's unit keys have already been updated.

A second reconciliation finding: the inherited package copier included the root
.fast-gate-lockfile-sha256 marker even though packages_digest excludes it.
copy_packages now omits that marker. A real copied-payload fixture and a mutant
which removes the copy exclusion prove that no unkeyed marker enters a product.
Directory enumeration by artifact.input_key's rglob may still visit excluded
physical install directories; these names supply no bytes to its digest.

The first shard trace also opened an existing ignored
`stage3/lane/__pycache__/check.cpython-312.pyc`. PYTHONDONTWRITEBYTECODE prevents
writes but does not prevent cached reads. The Go shard now sets
PYTHONPYCACHEPREFIX to its fresh results/pycache directory, with bytecode writes
disabled, so imports use source and cannot consume checkout bytecode. The final
trace proves the checkout bytecode read is absent. A forged valid checkout
check.pyc is ignored by the positive test; removing the private prefix causes
the real Go shard to consume it and fail with the planted marker. The initial
trace and its complete path list are retained separately. Empty/excluded cache
directory enumeration is structural only, like other excluded subtrees.

Both lane traces additionally open root `.gitattributes` during buildcache's
Git key bookkeeping. Declare that tracked path conservatively for all product
and key-test units and final-product fetch in the shard. Its sole current rule
is internal/oracle/testdata/reading/** -text, outside these inputs; it supplies
no lane recipe bytes, so its absence from Inputs.Files is explained rather than
silently overlooked. The companion apply inventory records it too.

## Wave 2 Go units

All 151 generated TestShard_ units have the same read categories as shard001
above, with their own manifest-selected tasks/cases in the fetched product.
`TestShardsMatchManifest` additionally reads `stage3/lane/generate_shards.py`,
`stage3/lane/shards.json` and `stage3/lane/*_test.go` for canonical generation and
AST namespace comparison. Go sources already belong to the Go closure.

`TestPythonUnitsMatchCommands` and all Python wrapper units read
`stage3/lane/python_units.py`, `stage3/lane/python-units.json`,
`stage3/test_apply.py`, `stage3/apply.py`, `stage3/caches.py`,
`stage3/lane/test_check.py`, `stage3/lane/test_table.py` and their imported lane
runtime from the table above. Discovery imports modules, but does not execute
cases. Apply shards read the declared apply product inputs and the fixtures'
source.json, API lock/pin/install, adaptation tree and test-shards.cjs. The
checkout proof reads its source checkout's tracked tree through git worktree
and compares ordinary apply output with that tree's patch-set.md. Preserve the
existing broad stage3 source read set for these units. Lane class tests also
read `stage3/adapt/75-optional-widening/evidence/full-apply/**`,
`stage3/adapt/32-indexed-reads-program/public-host-sites.json`, the sanctioned
API JSON, expected.json, normalize-api.cjs and oracle/observe-errors.cjs and
failure_details.py. All actual case IDs are committed in python-units.json.

`TestUnionIdentityOutcomes` reads `stage3/lane/test_union.py`,
`stage3/lane/merge_shards.py` and its imported check/shard_plan/caches modules.
Union measurement additionally reads the captured per-unit report.json,
observations.json, raw baseline diffs and whole-task profile outside the checkout;
these are explicit comparison inputs rather than repository globs. The source
and acceptance reads were audited against the new wrapper/bridge implementations;
the wave-1 empirical trace lists remain the shard/product runtime evidence.

Apply wrappers pass the fetched product99/products directory as
STAGE3_PRODUCT_STORE. The Python CLI reads and verifies the prepared manifest
and archive parts there, outside the source checkout; its read-only Go fallback
cannot build a missing product. This also works for the checkout-cleanliness
fixture's detached worktree without assuming root-specific local cache discovery.
