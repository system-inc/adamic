Defense blocked by red baseline

Starting commit: 6eef5ae586af0354d3032d36535e54cefb2da43a. nproc: 5. Warm toolchain used. npm ci completed before baseline.

Code under test: Adamic native regexp runtime and regexp compilation/native runner generation. Oracle: live Node 24 RegExp results, including captures, indices, named groups, and lastIndex. Tests also compare native test/exec agreement.

Both requested tests exist in the current test listing. Their whole source and shared checker were read. Search exercises 29 focused cases including Unicode surrogate-pair lastIndex boundaries; lint loads bench/regex/cases.json and exercises lint patterns and long inputs. Both assert actual matching observations against Node. Their names do not promise a performance threshold.

No mutants or coverage runs were performed: the mandatory clean baseline was red. TestDecodeASCIIUnit41 and TestDecodeASCIIUnit42 failed after about 55 seconds because decode-native subprocesses reported signal: killed. The package subsequently timed out at 90.05 seconds. The signal cause was not established. A timeout alone would allow narrowing, but these explicit failures preceded it.

Friction: the supplied audit command excerpts were truncated. Full prior evidence was fetched with the specified refspec. The current large package baseline runs parallel native subprocesses and did not finish within its budget. Go coverage cannot directly measure C runtime lines; no coverage claims are made.

Setup was skipped because the warm environment worked. npm timing was not separately measured. Baseline test-binary elapsed: 90.05 seconds, including failed rows at 55.06 and 55.07 seconds. No production rebuilds for mutants, no uniqueness matrix, no exclusive-line comparison, and no defense verdict against a green baseline were produced. Tests and production code were unchanged.
