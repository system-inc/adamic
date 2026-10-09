Built both exact-revision CLI compilers and emission probes; 19 programs, 38,139,115 C bytes identical.
Compared cut 3fefe5b221a751faa05cf49c865270c68c54b7d5 with baseline b3f83786a0ea00c23d47d774d3b6c98f9dd71336.
Both builds and final emission runs exit 0; whole-tree and every per-test diff -ru exit 0.
Emitted-comment mutant caught by diff exit 1; Unicode planted-failure leaf passes in 20.142s, asserting owning Node failure and neighbor success.
No compiler or test changes; no full gate, native execution or original semantic archive/CSS mutant reruns.

Evidence for release task #3tkk8dc: the C emission optimization preserves the red tests' compiled input programs byte for byte. This is an emission result, not a claim that their deadline failures are fixed.

| Test family | Programs | diff exit |
| --- | ---: | ---: |
| TestSixRuleAgreementAndMutants | 7 | 0 |
| TestTypeAwareAgreementAndMutants | 5 | 0 |
| TestVolumeConfigGuardAndMutant | 1 | 0 |
| TestVolumeAgreementCompiler | 1 | 0 |
| TestCompositionMatchesGoUnion | 1 | 0 |
| TestCSSPrinterAgreesWithGo | 4 | 0 |
| TestUnicodeNodeShardPlantedFailure | 0 | 0 |

Best of three native.C seconds, fresh IR each trial; largest means emitted C byte count.

| Program | C bytes | b3f83786 | 3fefe5b2 | baseline / cut |
| --- | ---: | ---: | ---: | ---: |
| volume-with-program-roots | 3724673 | 20.967974 | 4.700501 | 4.46x |
| guard-volume | 3719831 | 20.170922 | 4.590266 | 4.39x |
| printer | 2866675 | 9.203265 | 1.577574 | 5.83x |

nproc=5; cpu.max=400000/100000 (four CPU quota); GOMAXPROCS=4. Observed one-minute machine loads span 1.63 to 4.46; all three averages and active-task counts are retained per sample in timings.json. Some earlier samples overlap the identity baseline emitter, so ratios are observations at the recorded loads.

Setup timing lines: Go ready 0.064s, Node ready 0.080s, clang ready 0.448s, markdown dependency install step 1.247s, markdown ready 1.429s, submodules ready 17.999s, shared cache ready 24.158s. The setup command hit its 240s bound during cache warming; sourcing /workspace/adamic-tools/env.sh and building only the requested compilers/probes worked. Versions: Go 1.27.1, Node 24.19.0, clang 20.1.8.

Ancestry correction: 3fefe5b2 immediately descends from wrapper merge 1fe6d1bc; that merge has parents b3f83786 and 7b13a783. The comparison honors the two explicitly named compiler commits. TestVolumeAgreementCompiler is absent at the cut; its current main helper's generated program-roots adapter over the unchanged volume_suite.ts is included conservatively. Generated scratch Adamic inputs use .a. See coverage.md for every helper, source mutant and excluded runtime-only corpus/shard input.

Commands, build/emit logs, program/source/C SHA256s and all sample measurements are committed beside this report. No new fixtures or test leaves were added; counts.md and the call-target allowlist need no update. The emitted-comment mutation proves the byte comparator can fail; it is not an independent semantic oracle.

Integration lane checks passed: `lane checks 0.6 s: gofmt and tools on 0 Go files, t.Parallel on 0 test packages`. No changed Go/test packages needed vet or guard checks. Delivery branch: compiler/c-emission-identity.
