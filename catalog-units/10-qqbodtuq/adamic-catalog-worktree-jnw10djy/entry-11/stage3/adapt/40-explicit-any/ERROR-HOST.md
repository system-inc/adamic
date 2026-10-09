# Error host members

Sites 66, 67, 126 and 127 belong to Standard runtime feature probes.
Adaptation 40 replaces all four `(Error as any)` receivers with `Error`.
The unchanged compiler tsconfig selects `types: ["node"]`; the stock API
lockfile pins @types/node 25.3.3, whose ErrorConstructor already declares
captureStackTrace and stackTraceLimit. No host declaration, runtime check,
condition, argument, assignment, or public API sanction is added.

The adapter parses stock TypeScript 6.0.3, requires each exact reviewed source
line once, and replaces the parenthesized AsExpression receiver. It plans both
files before writing and tolerates line shifts from later adapters. A second
application changes zero bytes. Drift to an unknown assertion is rejected.
The historical classification ledger remains unchanged for traceability;
current progress is 74 of the original 210 any tokens removed, 136 remaining.

Stock TypeScript's transform erases the assertion and elides its redundant
parentheses around the identifier receiver. Thus `(Error as any).member`
already emits `Error.member`; removing both from the source emits exactly the
same bytes. The proof compares each entire debug.ts and sys.ts emission,
then every JavaScript and declaration artifact from clean upstream builds.
Source maps and build metadata are outside the byte-identity claim.

Reproduce with Node 24.19.0, the setup environment sourced, and NODE_PATH
pointing to the stock API node_modules cache:

```sh
bash stage3/apply.sh NEW_TREE > apply.log 2>&1
npm ci --prefix NEW_TREE --no-audit --no-fund > install.log 2>&1
node stage3/adapt/40-explicit-any/error-host-proof.cjs NEW_TREE NEW_PROOF > proof.log 2>&1
bash stage3/oracle/run.sh NEW_TREE NEW_ORACLE > oracle.log 2>&1
bash stage3/lane/run.sh NEW_LANE > lane.log 2>&1
```

The build proof reconstructs only these four original receivers, builds before
and after, then runs captureStackTraceWrong and stackTraceLimitWrong through
tsc's own clean build. Both must produce TS2551 or TS2339. It restores the
adapted tree and successfully rebuilds even on failure. Runtime statement
mutants change the capture argument or limit assignment and must change the
emitted JavaScript. An API byte mutant must fail the artifact comparison.

Temporary scanner adaptation 80's README.md, adapt.cjs and prove-js.cjs were
removed individually by name. The scanner profile no longer selects 80.
Gather a new slice from a stable fully adapted tree using the slice README's
createScanner, ScriptTarget and SyntaxKind entries. The profile is applied by
the slice tool and its 168 original spans are byte-audited. Run the documented
scanner driver with `--tree SLICE --inputs FIXED_CORPUS --oracle REFERENCE
--node-only`. Never gather while the build proof is mutating that tree.

The historical fixed corpus uses adaptations 00 and 10 before adaptation 50's
two explicit returns. Its absolute path is
`/workspace/scratch/scanner-fixed-inputs`, which is significant because pass
headers print paths. The initial default-driver corpus included both returns;
it had 1,369,441 tokens and was not the requested fixed corpus. Reconstructing
the pre-50 compiler inputs reproduces 1,369,432 tokens, 466 errors,
108,019,935 bytes and SHA-256
`41672da9bab56f9d10ad7d45b5938f96e3f969c6a299e5be1b260cba189893cc`.
Corpus hashes and both full-tree/slice reports are in evidence/error-host.
The driver's first token end+1 mutant is rejected by diff (exit 1).

Native scanner execution, parser-directed rescanning, Windows host behavior,
and the full Adamic Go gate are not claimed by this source adaptation.

Observed clean builds preserve all 10 JavaScript artifacts and 715 declarations
including the public typescript.d.ts. debug.ts's individual emission is 52,569
bytes; sys.ts's is 69,517 bytes. Both wrong-member build mutants produce TS2551
at both accesses, and the restored clean build exits 0. The seven initial mutant checks
pass; the final guard run additionally rejects duplicate lines, missing sites,
and renamed function owners for both files. exact artifact hashes are in evidence/error-host/proof.json.

Setup used GOPROXY=https://proxy.golang.org|direct and succeeded: Node 0.061s,
Go 0.072s, clang 0.489s, markdown dependencies 0.906s, submodules 380.152s,
Go build 611.985s, test binaries deferred 612.072s, cache warm 612.073s,
done 612.103s. nproc=5; CPU quota=4. The printed environment is
/workspace/adamic-tools/env.sh.

The full standalone oracle used NODE_OPTIONS=--max-old-space-size=2048 and
`--workers=2`. Install and build exit 0; tests exit 1 with 106,366 passing,
one failing, zero pending. The only failure and baseline difference are the
existing sanctioned public API acknowledgement and api/typescript.d.ts.
The initial four-worker lane ran during setup/build proofs and lost a worker;
its zero counts are not a pass. The container recorded one OOM kill. Completed
exploratory dumps were moved from tmpfs to workspace disk, and the final lane
runs separately with a bounded Node heap. Initial failure logs are preserved.

The final default four-worker lane reports PASS stage3 landing lane. It used
NODE_OPTIONS=--max-old-space-size=2048, a fresh tree on workspace disk, and
Node 24.19.0. Apply, install and build exit 0; the full suite has 106,366
passing, one sanctioned API acknowledgement failure and zero pending.
Only api/typescript.d.ts differs. No new OOM kill occurred. The final lane
reports, patch table, baseline diff and complete logs are in evidence/error-host.
Syntax checks and the staged git diff --check pass.
