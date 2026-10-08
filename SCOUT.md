# Parser scout corpus and result

Branch `parser-scout/census`, base `ad7bd06632f1` from origin/area/stage1-lint (recovery ancestor `88f4a83d`). Work is isolated in `/workspace/parser-scout`; no parser or shared checkout source was edited.

GitHub was reachable. All 23 supplied public repository snapshots were shallow-fetched at their exact SHAs with no installs, hooks, or repository scripts. The full quiet hundred lives on Kirk's Mac and was unavailable; these 23 snapshots are its supplied public subset. Explicit mappings for abbreviated repository names are in [public-pins.tsv](stage1/typescript/parser/census/public-pins.tsv). Every physical `.ts` and `.tsx` file, including declarations and malformed fixtures, was selected. No @filename virtual-file splitting was applied.

Corpus coverage:

| Corpus | Selected | Completed on both sides | Full byte identity | Identity excluding node flags | Parser failures | Adapter failures |
|---|---:|---:|---:|---:|---:|---:|
| 23 pinned public snapshots | 152,660 | 152,609 | 65,282 | 151,911 | 51 | 0 |
| TypeScript v6.0.3 src/compiler (parser gate) | 77 | 77 | 0 | 77 | 0 | 0 |
| stage3/drivers/tsc/corpus (tsc gate) | 300 | 300 | 202 | 298 | 0 | 0 |

The compiler files overlap the public TypeScript snapshot; these rows are not added together. The public corpus has 698 completed files differing beyond flags. The 51 parser failures replayed consistently: 41 JSX panics and 10 five-second parse limits. Census-only traces identify the repeated paths; no fix was attempted.

Oracle pins: cohere `7945d102a6c18dd36adf9114a758ce646e8b2359`, its TypeScript checkout `d92d9bfee114c80be2c375d72edae966176e3a4f`, and the gate compiler corpus `050880ce59e30b356b686bd3144efe24f875ebc8`. Go was `/workspace/adamic-tools/go/bin/go` (go1.27.1); Node v24.19.0. The port ran its actual TypeScript source on Node using the repository runtime hook. Native Adamic execution was not separately measured.

The ranked classes, shortest complete corpus inputs, full Go and port trees, function attribution, entry-point probes, and limitations are in [CENSUS.md](stage1/typescript/parser/CENSUS.md). [run.py](stage1/typescript/parser/census/run.py) is the comparison driver; [build.py](stage1/typescript/parser/census/build.py) builds the Go overlay, and [fetch.py](stage1/typescript/parser/census/fetch.py) fetches the pinned public inputs. Large inventories/results are gzip files: `fetch-results.json.gz`, `public.manifest.gz`, and `public.json.gz`. The source checkout paths in inventories preserve the run's `/workspace` layout.

The initial 77-file checkpoint was pushed as `e3947714`. This final extension is the second requested push. No third push, merge, or PR is part of this unit.
