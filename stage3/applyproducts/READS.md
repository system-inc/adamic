# Apply product and test reads

Planner inventory for Task: #hefhych. Paths are repository relative; `**`
includes every descendant, including dotfiles. This is a Loom unit read list,
not a substitute for buildcache.Inputs. Go sources and their Go import closure
are supplied by Loom already. Declare the union below for this package's units;
the narrower boundary lists describe each product precisely.

| Path or glob | Reader | Observation and buildcache reconciliation |
|---|---|---|
| `stage3/source.json` | All four products, all four TracksInputs tests; Fetch/Get | Successful open in cold product trace; explicit Inputs.Files. TracksInputs changes this pin in a private copy. |
| `stage3/apply.py`, `stage3/caches.py` | All products and TracksInputs tests | Successful Python source opens; explicit Inputs.Files. |
| `stage3/api/package.json`, `stage3/api/package-lock.json`, `stage3/api/node_modules.sha256` | All products and TracksInputs tests | Lock and pin opened in trace; package.json is conservatively declared even where not opened. Explicit Inputs.Files. |
| `stage3/api/node_modules/**` | All product workers, directly or through Node's parser resolution | Every file is read by packages_digest; observed opens include hidden lock, TypeScript JS/declarations, @types/node and undici-types. Physical install omitted from Inputs.Files intentionally: installed_packages checks the lock inventory and versions; pinned_packages checks every byte, executable bit, directory name and symlink against the declared api/node_modules.sha256. Loom supplies this NodeProject from its lockfile. |
| `stage3/adapt/00-setup/**`, `stage3/adapt/10-*/**` | Product 10 and its TracksInputs test | Directory opens and adapt.cjs execution observed. Inputs.Files declares the complete directories, including proofs and non-code files. |
| `stage3/adapt/[0-3][0-9]-*/**`, `stage3/adapt/40-*/**` | Product 40 and its TracksInputs test | Cumulative directory selection in inputs(); predecessor 10 fetched; worker also validates the parser. Source audit supplements product-10 trace. |
| `stage3/adapt/[0-6][0-9]-*/**`, `stage3/adapt/70-*/**` | Product 70 and its TracksInputs test | Cumulative Inputs.Files; predecessor 40 fetched. Source audit supplements trace. |
| `stage3/adapt/**` | Product 99 and its TracksInputs test; Fetch/Get | Cumulative Inputs.Files, including top-level adapt files. Final save_product/product_key reads the complete adaptation tree except __pycache__. Source audit supplements trace. |
| `go.mod` | All product/key operations | Successful root-discovery open. Go module metadata belongs to Loom's Go closure; not an additional product semantic input. Missing stage3/go.mod and stage3/applyproducts/go.mod probes also observed. |

The deadline test reads no repository data beyond module-root discovery and
Go sources. It executes inline Python and Node and reads /proc in its private
process proof. TracksInputs tests copy all declared paths, preserving modes
and symlinks; their newly written source.json is private test output. The
fetch command has the same read set as Get, rather than a second recipe.

## Empirical evidence and exceptional reads

`evidence/planner-reads/apply-paths.tsv.gz` lists every repository path opened
or probed by the traced cold product, with read/write/exec/probe classification.
The corresponding complete syscall trace and the parser are archived beside it.
The trace command, exit, timing and the untraced cold measurements are recorded
in PLANNER_FIXES_REPORT.md. It runs the compiled Go test binary so compiler reads
are not confused with product-time reads. This is an empirical lower bound;
source audit of all four recipes and all tests supplies conditional paths above.

Git's index acceleration also opens `.git/**`, `.gitmodules`, `cohere/.git`,
`cohere/.gitmodules` and `cohere/TypeScript/.git`, including submodule object
packs and attempted sequencer/config paths. These are buildcache's generic
Git bookkeeping, not compiler inputs. Their content is not an undeclared
adaptation dependency: buildcache values each declared path by blob and mode.
Do not turn mutable index files or logs into semantic read lines. The complete
path listing retains them so this exclusion is reviewable.

npm's `--version` probes a package-local `.npmrc` (absent); Python probes an
absent caches __pycache__ entry. Neither supplies bytes. Products do not write
bytecode. No checkout-local npm config is part of these recipes; a future one
would need declaring or refusing. Installed tool executables, Python/Node/npm
standard libraries, OS configuration and temporary/cache/product directories
are outside the repository. Tool version reports are in Inputs.Toolchain.

Product 10 reads the gate-supplied pinned TypeScript checkout outside this
repository (`ADAMIC_TYPESCRIPT_SOURCE`), verifies its commit against source.json,
and clones it via file://. Later products read independent predecessor product
copies. These are source-pin and product dependencies, not repository globs.
No source or package download belongs to a product.

Inputs.Files includes Go recipe source and conservative proof files that a
particular successful worker does not open. That is intentional over-keying.
The observed non-Go semantic repository reads are covered by the table: directly
by Inputs.Files or by the validated API byte pin. Loom must additionally key
these reads for test results, even when buildcache already keys a product.

Reconciliation finding: product 99 calls product_key(), which also reads
`stage3/adapt/CHECKER-OPTIONS.md` and `stage3/adapt/checker-options-evidence.json`.
The old Go recipe declared only adaptation directories. Product 99 now declares
every adaptation entry, including those two files. TestStage3Adapted99TracksMetadata
changes the real top-level metadata in a private copy; omitting those entries
is its input-drop mutant. This finding came from auditing the conditional final
archive path after reconciling the product-10 trace.
