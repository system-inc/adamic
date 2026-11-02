# #wd9wpyk: the latent census's record answers once a file

The commit this branch sits on (compiler/latent-entries-once) shares entriesRecordLiteral's memo and
entriesProgramHasRecords' answer across every lowering of a file. These are the measurements behind its message.

The box had 4 cores and 15 GB, ran Linux, and had nothing else running. Inputs: tsc v6.0.3 (050880ce) adapted by main's
`stage3/apply.sh` (79 files under src/compiler). Binaries came from `harness/build.sh`, the meter's `build_census`
for the latent binary, and each run used `harness/run.sh`, twice-daily.sh line 113's latent step (LATENT_FULL,
LATENT_ASSERT_NO_OUTPUT, LATENT_TIMES), timed with getrusage of the child. "Before" is origin/main 80c2d328.

## The whole program doesn't fit on this box

`runs/out-before`: main's binary over all 79 files was killed by the OOM killer (`oom.txt`) at 13.8 GB RSS, after 17
files and 3009 s. Its latent.jsonl holds only the header, so its sha means nothing. So I didn't measure the whole
program. Instead, a scratch-only `LATENT_ONLY_FILES` list filter (`harness/subset_units.go.txt` for latentFull and
`harness/subset_hook.go.txt` for latentRun, each main's generated overlay plus that one `if`) was laid identically into
every binary. It ran six files that the full run had reached: binder, builder, commandLineParser, debug,
executeCommandLine and expressionToTypeNode.

| run | mode | wall s | user s | latent.jsonl sha256 |
|---|---|---|---|---|
| sub-before | full | 613.1 | 782.0 | eb629e30...5dcf |
| sub-after | full | 345.2 | 502.9 | eb629e30...5dcf |
| sub-across | full, one share for the run | 316.3 | 468.7 | eb629e30...5dcf |
| sub-across-without-statecopy | full, across variant built without the statecopy pass-through (my mistake, kept) | 637.0 | 798.3 | eb629e30...5dcf |
| first-before | first-error | 225.9 | 252.9 | 562d86e9...c2de |
| first-after | first-error | 126.5 | 151.7 | 562d86e9...c2de |
| etn-before | full, expressionToTypeNode.ts alone (LATENT_ONLY_FILE) | 83.1 | 99.8 | a77317da...3e1d |
| etn-after | same | 63.5 | 81.0 | a77317da...3e1d |
| one-debug-before / after | debug.ts alone, after = the program answer shared without the memo | 33.7 / 34.2 | 42.2 / 43.1 | 365a00c0...8326 |

latent.jsonl carries no timing fields (those go to latent-times.jsonl), so I compared the shas as they are.

The across variant (`harness/across-files-variant-vs-commit.*.diff`) was measured and not shipped. Whether another
file's checker can change an answer is Lower's to say.

## Profiles (expressionToTypeNode.ts, full mode)

- `profiles/expressionToTypeNode-before.*`: entriesProgramRecordsWalk is 12.35 s of 101.18 s, all under
  entriesProgramHasRecords, which is called only from property (object.go:424 and :657), which enumNeverValue
  dispatches to (`*.enumNeverValue-peek.txt`). That walk is enumNeverValue's program-wide cost. Its answer is a
  function of the program and the file's checker (entriesRecordLiteral's own comment, and moduleOrder of Files()[0]),
  kept per lowering in entriesProgramRecords/Known.
- `profiles/expressionToTypeNode-program-answer-only.*`: the brief's first form, with only entriesProgramRecords/Known
  shared. The walk disappears and entriesRecordStorageTypes pays the same literal walks (19.15 s), because the program
  walk had been filling the per-literal memo it reads. The wall didn't move (84.7 s), which is why the commit shares
  the memo too.
- `profiles/expressionToTypeNode-after.*`: entriesProgramRecordsWalk is 2.46 s of 82.22 s, and entriesRecordLiteral
  4.56 s, which is one walk a literal a file. What's left at the top is latentSnapshot (14.5 s), predicate proofs and
  moduleOrder (4.1 s, asked by many callers per unit). None of those are in this change.

## Mutants (`mutants/`)

Each was planted in the working tree, run after its products were rebuilt, and then restored:

- share-line-removed: lower.go.txt's latentRun lost `latentEntriesShare = latentEntriesFile()`. Caught by
  TestEachFileStartsAFreshShare: "a lowering is made before the file's share is set".
- full-without-literals: units.go.txt's fresh() lost `entriesRecords: latentEntriesShare.literals`. Caught by
  TestEachFileStartsAFreshShare: "a lowering is made without entriesRecords".
- share-write-dropped: `share.known, share.records = true, records` became `_ = share`. Caught by TestOneWalkAFile
  through its probe: "file 0: the second lowering's answer isn't the file's first walk's".
- statecopy-case-dropped: the generator's entriesRecords case became `false &&`. Caught by
  TestLatentStatecopySharesEntriesRecords: "the lowering copier deep-copies entriesRecords".

## Not done

- No whole-program before/after (OOM, above). The speculative census has no share and was not run.
- Byte equality is shown on the six-file subset and two single files, in both meter modes, not on all 79 files.
