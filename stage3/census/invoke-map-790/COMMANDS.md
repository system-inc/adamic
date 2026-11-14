# Commands and scope

Scratch root: `/tmp/invoke-map-4pc8wd8`. All commands ran in its detached
`repo` worktree unless a directory is shown. The shared checkout supplied Git
objects/worktree metadata only. No Loom command or submission was invoked.

Setup:

```sh
git -C /workspace/adamic fetch origin 8e3a784aa5a601f57342c07b68f145fe56252c21
git -C /workspace/adamic worktree add --detach /tmp/invoke-map-4pc8wd8/repo 8e3a784aa5a601f57342c07b68f145fe56252c21
git submodule update --init --recursive --depth 1
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /tmp/invoke-map-4pc8wd8/TypeScript
(cd stage3/api && npm ci)
source /tmp/invoke-map-4pc8wd8/env.sh
```

`env.sh` records scratch cache/source paths. Source HEAD was confirmed as
050880ce59e30b356b686bd3144efe24f875ebc8. Node v24.19.0, Go go1.27.1 linux/amd64.
The implementation commit is a1f279e55e39e06d450da64d413d21f508017396, whose sole
parent is the exact requested base.

Measurement build:

```sh
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/invoke-map-4pc8wd8/overlay
go build -buildvcs=false -overlay=/tmp/invoke-map-4pc8wd8/overlay/overlay.json -o /tmp/invoke-map-4pc8wd8/latent-census ./stage3/census/latent/tool
go build -buildvcs=false -o /tmp/invoke-map-4pc8wd8/adamic ./cmd/adamic
python3 stage3/census/latent/audit.py /tmp/invoke-map-4pc8wd8/latent-census
python3 stage3/census/latent/full_audit.py /tmp/invoke-map-4pc8wd8/latent-census /tmp/invoke-map-4pc8wd8/evidence/census-audit
```

Both builds and the legacy audit exit 0. The supplemental full audit exits 1
on its obsolete condition-refusal expectation; full raw fixture observations
are retained. Production sources and overlay tooling were not edited.

Main, amendment, and restoring-mutant measurements use the same binary and
source path, sequentially, with precisely the one declaration exchanged. Their
full source versions are `main-sys.ts` and `amendment-sys.ts`; after the mutant,
the amendment is restored. Each command below exits 0. OUTPUT is respectively
main-census.jsonl, amendment-census.jsonl, and restoring-mutant-census.jsonl.

```sh
CENSUS_MODULE_FORMAT=CommonJSBundle LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 \
LATENT_ONLY_FILE=/tmp/invoke-map-4pc8wd8/main-tree/src/compiler/sys.ts \
/tmp/invoke-map-4pc8wd8/latent-census /tmp/invoke-map-4pc8wd8/main-tree/src/compiler \
/tmp/invoke-map-4pc8wd8/evidence/OUTPUT
python3 /tmp/invoke-map-4pc8wd8/compare.py
```

The comparator preserves kind, exact where/reason/text, phase, unit and boundary
spans. It maps sys.ts byte offsets after the shorter annotation back to main's
coordinates; offsets in dependency files are never changed. The entire raw
restoring-mutant stream is independently required to equal main byte-for-byte.
The census is an observation ledger with isolated attempts and its usual
coverage limits; it does not claim successful native compilation.

Stock checking and JavaScript emission, on each side before mutating source:

```sh
(cd /tmp/invoke-map-4pc8wd8/main-tree && npm ci --no-audit --no-fund)
node /tmp/invoke-map-4pc8wd8/stock.cjs /tmp/invoke-map-4pc8wd8/main-tree /tmp/invoke-map-4pc8wd8/evidence/SIDE-stock
/tmp/invoke-map-4pc8wd8/adamic types $(find /tmp/invoke-map-4pc8wd8/main-tree/src/compiler -name '*.ts' | sort)
(cd /tmp/invoke-map-4pc8wd8/main-tree && ./node_modules/.bin/hereby dts)
```

All exit 0. `stock.cjs` uses the repository's existing strict compiler-project
helper for diagnostics, enumerates all 79 recursive roots, and separately emits
JavaScript. Every file's actual bytes are compared, with no normalization.
The emitted files are retained in SIDE-stock-javascript.tar.gz and per-file
hashes in SIDE-stock/emitted.json. Upstream's actual hereby dts outputs are
retained under SIDE-dts and compared byte-for-byte. The amendment declaration
build follows the main build in the same tree and reruns its source-dependent
build tasks. No hand-written declaration substitute is used.

Existing whole-owner proof:

```sh
node stage3/adapt/99-correctness-timer-closures/contract.cjs /tmp/invoke-map-4pc8wd8/contract-input
node stage3/adapt/99-correctness-timer-closures/witness.cjs /tmp/invoke-map-4pc8wd8/main-tree
```

The contract fixture contains every exact before-pinned owner grouped by its
original file; the same fixture is used with main's and the amendment's owners.
Both contracts and both actual adapted-tree witnesses exit 0; each stdout pair
is byte-identical. The initial pristine-upstream contract invocation was an
incorrect input and failed; it was replaced by this exact before-pin fixture.

Apply and lane:

```sh
stage3/apply.sh /tmp/invoke-map-4pc8wd8/main-tree
stage3/lane/run.sh /tmp/invoke-map-4pc8wd8/lane
```

The initial main apply finished all adaptations, then exited 1 during cache tar
creation because proof builds were started at patch-set creation before tar
finished. Its failed receipt is retained. The amendment lane is run whole,
without --artifact, filters, shard substitutions, or edits to lane/oracle.
Nothing touches its adapted tree until its apply process returns.
Its final execution/report/verdict and raw phase logs are retained separately.
