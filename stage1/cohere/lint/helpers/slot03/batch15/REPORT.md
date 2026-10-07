Built pushJump, popJump and makeUnreachable in separate .a files: twelve dependency edges across four rules.
Commits: claim e4b23dd4 pushed before source; implementation 39f8aa679c28149cca72ec80d95713e385c0c517 rebased onto advanced main b8fb957a; publication only to codex/lint-helpers-03.
Commands: focused Go/source Node/emitted JavaScript/sanitized native PASS 59.403s, 6,997 snapshots; current-main owned gate PASS 584.023s, 3,041,759 lines; inherited helpers PASS 74.785s; vet/format clean; six uncached probes PASS 1.739s.
Mutants: eleven transition mutants and an empty-stack guard mutant compile and are caught; all 146 owned variants, four inherited variants and missing-consumer check caught.
Not covered: complete native CFG, dependency wiring, whole-rule findings/fixes/suggestions or full repository gate.

## Claims and landing

Only codex/lint-helpers-03 is pushed by this worker. All forty-two earlier helpers were complete, oracle-green and pushed at a2519dfd on c01907a before claiming. Every claims file across twenty origin/codex/lint-helpers* branches was checked. The comments bundle remains reserved by shared HELPERS.md. These concrete CFG helpers tie the highest available count at four. Regex-engine internals are not selected for a handwritten matcher. Claim e4b23dd4 precedes source.

A refreshed scan finds later duplicate pushJump/popJump claims on slot 05. Slot 03 e4b23dd4 at 06:00:46 UTC precedes slot 05 85af6690 at 06:02:56 UTC, so slot 03 retains them. makeUnreachable remains unique. During validation main advances from c01907a to b8fb957a, adding inherited static-field reads. The old-base full gate is stopped for rebase; it is not counted as a pass. The user explicitly authorizes rebase onto current main. No main or area branch is pushed.

## Behavior and independent evidence

README.md documents the three contracts. Node/block identities are opaque numeric handles (-1 for nil), not numeric AST kinds. These helpers do no kind dispatch. LabelsOf, block allocation and successor storage remain explicit separately owned dependencies. The private builder stack does not expose Go slice-header aliases; replacement arrays preserve target-object identities, but arbitrary Go backing-store aliases are outside this interface. Barrier presence and its values are separate fields because stage 0 refuses a boolean-array/null field; no compiler change is made.

The oracle uses actual Go CFG Build on every parser root of the captured complete runtime sources. Temporary overlays rename the three helper method definitions only to install before/after wrappers; helper bodies remain unchanged. Actual Go supplies label lists, block handles and before/after snapshots. Native/source/emitted adapters replay each operation from its observed pre-state using explicit dependencies. This proves these state transitions, not native graph traversal or the separately owned dependencies.

The source capture covers all four consumers and 2,119 runtime inputs. Upstream core/react package capture gate exits 0. Controls add nested labels, do/for/switch/break/continue/return/throw/finally roots, stack depths 0..7, successor counts 0..9, both reachability/incoming flags and absent/present barrier arrays. There are 6,997 snapshot lines. Empty-pop rejection is observed separately in real Go; source Node, emitted JavaScript and native panic with exit 70 and the owned invariant message. Its guard mutant compiles, exits 0 without stderr, and prints survived. Invalid ASTs beyond parser-produced consumer roots and arbitrary cyclic graph representations are not covered.

Initial oracle failure: relative fixture filenames caused the Go parser's absolute-path invariant panic. The owned adapter now normalizes filenames. A second failed run recorded stage 0's unsupported boolean[] | null field. The port's explicit presence flag resolves it. evidence/focused.log and focused-retry.log retain these failures; focused-supported.log is the successful transition gate, and empty-guard.log is the successful refusal mutant gate.

## Consumers

Each helper removes one dependency from each of array-callback-return, consistent-return, no-unreachable-loop and react-hooks/rules-of-hooks. Twelve edges, four distinct rules; readiness.json subtracts only this worker's forty-five delivered helpers, never other workers' claims. It lists the remaining dependencies and any final-helper readiness. No native whole-rule parity is claimed.

## Commands

Setup `bash cloud/setup.sh`: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 44s, done 44s. nproc=5, cpu.max=400000 100000, 17.6GB. Go1.27.1, clang20.1.8, Node24.19.0. Source /workspace/adamic-tools/env.sh in each toolchain shell.

Focused transition gate: `go test ./stage1/cohere/lint/helpers/slot03 -run TestBatch15 -count=1 -v -timeout=20m` -> focused-supported.log, PASS 59.403s. Separate `-run TestBatch15EmptyStackRefusal` -> empty-guard.log, PASS 0.743s. Full touched-package gate: `go test ./stage1/cohere/lint/helpers ./stage1/cohere/lint/helpers/slot03 -count=1 -v -timeout=20m`. Repository vet `go vet ./...`; format `gofmt -l cmd internal stage1/cohere/lint/helpers/slot03`; six uncached input probes `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m`. Each command writes directly to an owned evidence log. Corpus regeneration runs twice; hashes and current-main results are recorded before publication.

## Every new transition mutant

- TestBatch15Mutants/push_jump.a: batch15_test.go:125: compiled semantic mutant caught at line 90: got "1/2:1:0:false:false:/labels:0;" Go "1/2:1:0:true:false:/labels:0;"
- TestBatch15Mutants/push_jump.a#01: batch15_test.go:125: compiled semantic mutant caught at line 90: got "1/2:1:0:true:true:/labels:0;" Go "1/2:1:0:true:false:/labels:0;"
- TestBatch15Mutants/push_jump.a#02: batch15_test.go:125: compiled semantic mutant caught at line 90: got "1/2:1:0:true:false:/labels:1;" Go "1/2:1:0:true:false:/labels:0;"
- TestBatch15Mutants/push_jump.a#03: batch15_test.go:125: compiled semantic mutant caught at line 90: got "1/1:2:0:true:false:/labels:0;" Go "1/2:1:0:true:false:/labels:0;"
- TestBatch15Mutants/push_jump.a#04: batch15_test.go:125: compiled semantic mutant caught at line 90: got "1/2:1:-1:true:false:/labels:0;" Go "1/2:1:0:true:false:/labels:0;"
- TestBatch15Mutants/pop_jump.a: batch15_test.go:125: compiled semantic mutant caught at line 125: got "0/" Go "1/2:1:6:true:false:/"
- TestBatch15Mutants/pop_jump.a#01: batch15_test.go:125: compiled semantic mutant caught at line 92: got "1/2:1:0:true:false:/" Go "0/"
- TestBatch15Mutants/make_unreachable.a: batch15_test.go:125: compiled semantic mutant caught at line 6770: got "0:false:false/1,/true,|1:false:false//nil" Go "0:false:false/1,/false,|1:false:false//nil"
- TestBatch15Mutants/make_unreachable.a#01: batch15_test.go:125: compiled semantic mutant caught at line 1: got "2:true:true/2,/nil|3:false:false//nil" Go "2:true:true/3,/nil|3:false:false//nil"
- TestBatch15Mutants/make_unreachable.a#02: batch15_test.go:125: compiled semantic mutant caught at line 1: got "2:true:true/3,/nil|2:true:true/3,/nil" Go "2:true:true/3,/nil|3:false:false//nil"
- TestBatch15Mutants/make_unreachable.a#03: batch15_test.go:125: compiled semantic mutant caught at line 1: got "2:true:true/3,/nil|3:true:false//nil" Go "2:true:true/3,/nil|3:false:false//nil"

Empty-stack guard deletion is caught by the exit-70 refusal comparison; its compiled mutant exits 0 and prints survived. All transition variants must compile and execute with empty stderr before changed snapshots count. Full current-main regression records earlier mutants too.

## Current-main landing and every regression witness

The rebase onto b8fb957aa839a9e8cb0b54279dd9864fa317bd30 completes cleanly across 57 branch commits; the owned implementation is unchanged by rebase. The previous partial old-base gate exits 143 when stopped, and is not a pass. Current-main logs are final.log, vet-current-main.log (empty), format-current-main.log (empty) and input-current-main.log (six probe misses, zero hits). All forty-five helpers are re-green on this base. Shared harness ab70f38d4 and allocator-check changes are not yet on this main; no shared change is reverted.

Corpus hashes reproduce byte-for-byte: identical regeneration: {"coverage.json": "1f0ca1607de0ad46eccf3d5d477200efb20c0e27be871f93af8b547110740115", "sources.jsonl.gz": "a1693be8bd8b2e385e79206d208166de3608b9d325e7aa8b2e96f22683824edf"}. Upstream consumer package capture exits 0. Native whole-rule findings, complete native CFG traversal, arbitrary externally aliased Go slice headers and separately owned dependency wiring remain outside coverage.

Slot 05 explicitly withdraws pushJump/popJump in its claim file at origin/codex/lint-helpers-05 tip 2aa00876; slot 03 retains the earliest published claim. No duplicate implementation from this unit and no fourth claim.

- TestHelperMutants/options_json.ts: helpers_test.go:145: compiled semantic mutant caught at output line 15157: got "valid", Go "invalid"
- TestHelperMutants/option_schema.ts: helpers_test.go:145: compiled semantic mutant caught at output line 15166: got "valid", Go "invalid"
- TestHelperMutants/policy_message.ts: helpers_test.go:145: compiled semantic mutant caught at output line 22166: got "This throws a bare `{{constructor}}`, which names no declared failure. Raise it through the tier that declares it, `AccountModule.error(identifier, data, cause)`, `ApiWorker.error(...)` or `Base.error(...)`. A bare throw carries no identifier, so the board groups it by its message and one interpolated value mints one identity per value, and it normalizes to 500, so a refusal reads as our fault.", Go "This throws a bare `sentinel é😀`, which names no declared failure. Raise it through the tier that declares it, `AccountModule.error(identifier, data, cause)`, `ApiWorker.error(...)` or `Base.error(...)`. A bare throw carries no identifier, so the board groups it by its message and one interpolated value mints one identity per value, and it normalizes to 500, so a refusal reads as our fault."
- TestHelperMutants/strict_options.ts: helpers_test.go:145: compiled semantic mutant caught at output line 15178: got "valid", Go "invalid"
- TestBatch10Mutants/is_url.a: batch10_test.go:91: compiled semantic mutant caught at line 79: got "true" Go "false"
- TestBatch10Mutants/is_url.a#01: batch10_test.go:91: compiled semantic mutant caught at line 916: got "true" Go "false"
- TestBatch10Mutants/is_absolute_size.a: batch10_test.go:91: compiled semantic mutant caught at line 1046: got "false" Go "true"
- TestBatch10Mutants/is_absolute_size.a#01: batch10_test.go:91: compiled semantic mutant caught at line 65: got "true" Go "false"
- TestBatch10Mutants/is_relative_size.a: batch10_test.go:91: compiled semantic mutant caught at line 783: got "false" Go "true"
- TestBatch11Mutants/is_generic_name.a: batch11_test.go:93: compiled semantic mutant caught at line 549: got "false" Go "true"
- TestBatch11Mutants/is_generic_name.a#01: batch11_test.go:93: compiled semantic mutant caught at line 101: got "true" Go "false"
- TestBatch11Mutants/has_math_function.a: batch11_test.go:93: compiled semantic mutant caught at line 38: got "false" Go "true"
- TestBatch11Mutants/has_math_function.a#01: batch11_test.go:93: compiled semantic mutant caught at line 60: got "false" Go "true"
- TestBatch11Mutants/has_math_function.a#02: batch11_test.go:93: compiled semantic mutant caught at line 88: got "true" Go "false"
- TestBatch11Mutants/loaded_utilities.a: batch11_test.go:93: compiled semantic mutant caught at line 1115: got "-1/false/true" Go "0/true/true"
- TestBatch11Mutants/loaded_utilities.a#01: batch11_test.go:93: compiled semantic mutant caught at line 1115: got "0/true/false" Go "0/true/true"
- TestBatch12Mutants/is_angle.a: batch12_test.go:120: compiled semantic mutant caught at line 1: got "false/suffix:deg,rad,grad;" Go "false/suffix:deg,rad,grad,turn;"
- TestBatch12Mutants/is_angle.a#01: batch12_test.go:120: compiled semantic mutant caught at line 1: got "true/suffix:deg,rad,grad,turn;" Go "false/suffix:deg,rad,grad,turn;"
- TestBatch12Mutants/is_number.a: batch12_test.go:120: compiled semantic mutant caught at line 2: got "true/scan;" Go "false/scan;math;"
- TestBatch12Mutants/is_number.a#01: batch12_test.go:120: compiled semantic mutant caught at line 92: got "true/scan;" Go "false/scan;math;"
- TestBatch12Mutants/is_number.a#02: batch12_test.go:120: compiled semantic mutant caught at line 2: got "false/math;scan;math;" Go "false/scan;math;"
- TestBatch12Mutants/is_percentage.a: batch12_test.go:120: compiled semantic mutant caught at line 3: got "false/suffix:percent;math;" Go "false/suffix:%;math;"
- TestBatch12Mutants/is_percentage.a#01: batch12_test.go:120: compiled semantic mutant caught at line 3: got "false/suffix:%;" Go "false/suffix:%;math;"
- TestBatch12Mutants/is_percentage.a#02: batch12_test.go:120: compiled semantic mutant caught at line 3: got "false/math;suffix:%;math;" Go "false/suffix:%;math;"
- TestBatch12ArgumentMutants/is_angle.a: batch12_test.go:166: compiled semantic mutant caught at line 1: got "false/suffix:deg,rad,grad,turn;wrong-value;" Go "false/suffix:deg,rad,grad,turn;"
- TestBatch12ArgumentMutants/is_number.a: batch12_test.go:166: compiled semantic mutant caught at line 2: got "false/scan;wrong-value;math;" Go "false/scan;math;"
- TestBatch12ArgumentMutants/is_number.a#01: batch12_test.go:166: compiled semantic mutant caught at line 2: got "false/scan;math;wrong-value;" Go "false/scan;math;"
- TestBatch13Mutants/matches_data_type.a: batch13_test.go:228: compiled semantic mutant caught at line 1: got "false/check:1;" Go "false/check:0;"
- TestBatch13Mutants/matches_data_type.a#01: batch13_test.go:228: compiled semantic mutant caught at line 2: got "false/check:2;" Go "false/check:1;"
- TestBatch13Mutants/matches_data_type.a#02: batch13_test.go:228: compiled semantic mutant caught at line 3: got "false/check:3;" Go "false/check:2;"
- TestBatch13Mutants/matches_data_type.a#03: batch13_test.go:228: compiled semantic mutant caught at line 4: got "false/check:4;" Go "false/check:3;"
- TestBatch13Mutants/matches_data_type.a#04: batch13_test.go:228: compiled semantic mutant caught at line 5: got "false/check:5;" Go "false/check:4;"
- TestBatch13Mutants/matches_data_type.a#05: batch13_test.go:228: compiled semantic mutant caught at line 6: got "false/check:6;" Go "false/check:5;"
- TestBatch13Mutants/matches_data_type.a#06: batch13_test.go:228: compiled semantic mutant caught at line 7: got "false/check:7;" Go "false/check:6;"
- TestBatch13Mutants/matches_data_type.a#07: batch13_test.go:228: compiled semantic mutant caught at line 8: got "false/check:8;" Go "false/check:7;"
- TestBatch13Mutants/matches_data_type.a#08: batch13_test.go:228: compiled semantic mutant caught at line 9: got "false/check:9;" Go "false/check:8;"
- TestBatch13Mutants/matches_data_type.a#09: batch13_test.go:228: compiled semantic mutant caught at line 10: got "false/check:10;" Go "false/check:9;"
- TestBatch13Mutants/matches_data_type.a#10: batch13_test.go:228: compiled semantic mutant caught at line 11: got "true/check:11;" Go "false/check:10;"
- TestBatch13Mutants/matches_data_type.a#11: batch13_test.go:228: compiled semantic mutant caught at line 12: got "false/check:12;" Go "true/check:11;"
- TestBatch13Mutants/matches_data_type.a#12: batch13_test.go:228: compiled semantic mutant caught at line 13: got "false/check:13;" Go "false/check:12;"
- TestBatch13Mutants/matches_data_type.a#13: batch13_test.go:228: compiled semantic mutant caught at line 14: got "false/check:14;" Go "false/check:13;"
- TestBatch13Mutants/matches_data_type.a#14: batch13_test.go:228: compiled semantic mutant caught at line 15: got "false/check:15;" Go "false/check:14;"
- TestBatch13Mutants/matches_data_type.a#15: batch13_test.go:228: compiled semantic mutant caught at line 16: got "false/check:16;" Go "false/check:15;"
- TestBatch13Mutants/matches_data_type.a#16: batch13_test.go:228: compiled semantic mutant caught at line 17: got "false/check:0;" Go "false/check:16;"
- TestBatch13Mutants/matches_data_type.a#17: batch13_test.go:228: compiled semantic mutant caught at line 18: got "false/check:0;" Go "false/"
- TestBatch13Mutants/matches_data_type.a#18: batch13_test.go:228: compiled semantic mutant caught at line 1: got "false/check:0;wrong-value;" Go "false/check:0;"
- TestBatch13Mutants/infer_data_type.a: batch13_test.go:228: compiled semantic mutant caught at line 29879: got "/type:color;check:0;type:length;check:1;type:percentage;check:2;type:ratio;check:3;type:number;check:4;type:integer;check:5;type:url;check:6;type:position;check:7;type:bg-size;check:8;type:line-width;check:9;type:image;check:10;type:family-name;check:11;type:generic-name;check:12;type:absolute-size;check:13;type:relative-size;check:14;type:angle;check:15;type:vector;check:16;" Go "/"
- TestBatch13Mutants/infer_data_type.a#01: batch13_test.go:228: compiled semantic mutant caught at line 14279: got "/" Go "length/type:color;check:0;type:length;check:1;"
- TestBatch13Mutants/infer_data_type.a#02: batch13_test.go:228: compiled semantic mutant caught at line 23: got "family-name/type:vector;check:16;type:angle;check:15;type:relative-size;check:14;type:absolute-size;check:13;type:generic-name;check:12;type:family-name;check:11;" Go "family-name/type:color;check:0;type:length;check:1;type:percentage;check:2;type:ratio;check:3;type:number;check:4;type:integer;check:5;type:url;check:6;type:position;check:7;type:bg-size;check:8;type:line-width;check:9;type:image;check:10;type:family-name;check:11;"
- TestBatch13Mutants/infer_data_type.a#03: batch13_test.go:228: compiled semantic mutant caught at line 23: got "family-name/type:color;wrong-value;check:0;wrong-value;type:length;wrong-value;check:1;wrong-value;type:percentage;wrong-value;check:2;wrong-value;type:ratio;wrong-value;check:3;wrong-value;type:number;wrong-value;check:4;wrong-value;type:integer;wrong-value;check:5;wrong-value;type:url;wrong-value;check:6;wrong-value;type:position;wrong-value;check:7;wrong-value;type:bg-size;wrong-value;check:8;wrong-value;type:line-width;wrong-value;check:9;wrong-value;type:image;wrong-value;check:10;wrong-value;type:family-name;wrong-value;check:11;wrong-value;" Go "family-name/type:color;check:0;type:length;check:1;type:percentage;check:2;type:ratio;check:3;type:number;check:4;type:integer;check:5;type:url;check:6;type:position;check:7;type:bg-size;check:8;type:line-width;check:9;type:image;check:10;type:family-name;check:11;"
- TestBatch13Mutants/is_family_name.a: batch13_test.go:228: compiled semantic mutant caught at line 5328: got "true/segment:,;" Go "false/segment:,;"
- TestBatch13Mutants/is_family_name.a#01: batch13_test.go:228: compiled semantic mutant caught at line 14304: got "false/segment:,;" Go "true/segment:,;"
- TestBatch13Mutants/is_family_name.a#02: batch13_test.go:228: compiled semantic mutant caught at line 48: got "false/segment:,;" Go "true/segment:,;"
- TestBatch13Mutants/is_family_name.a#03: batch13_test.go:228: compiled semantic mutant caught at line 29904: got "true/segment:,;" Go "false/segment:,;"
- TestBatch13Mutants/is_family_name.a#04: batch13_test.go:228: compiled semantic mutant caught at line 48: got "true/segment:;;" Go "true/segment:,;"
- TestBatch13Mutants/is_family_name.a#05: batch13_test.go:228: compiled semantic mutant caught at line 48: got "true/segment:,;wrong-value;" Go "true/segment:,;"
- TestBatch14Mutants/is_line_width.a: batch14_test.go:135: compiled semantic mutant caught at line 955: got "false/segment: ;length:0;number:0;length:1;number:1;" Go "true/segment: ;length:0;number:0;length:1;number:1;"
- TestBatch14Mutants/is_line_width.a#01: batch14_test.go:135: compiled semantic mutant caught at line 1: got "false/segment: ;number:0;length:0;number:0;" Go "false/segment: ;length:0;number:0;"
- TestBatch14Mutants/is_line_width.a#02: batch14_test.go:135: compiled semantic mutant caught at line 1: got "false/segment: ;length:-1;number:0;" Go "false/segment: ;length:0;number:0;"
- TestBatch14Mutants/is_line_width.a#03: batch14_test.go:135: compiled semantic mutant caught at line 1: got "false/segment: ;wrong-value;length:0;number:0;" Go "false/segment: ;length:0;number:0;"
- TestBatch14Mutants/is_line_width.a#04: batch14_test.go:135: compiled semantic mutant caught at line 1: got "false/segment:,;length:0;number:0;" Go "false/segment: ;length:0;number:0;"
- TestBatch14Mutants/is_line_width.a#05: batch14_test.go:135: compiled semantic mutant caught at line 1: got "true/segment: ;length:0;number:0;" Go "false/segment: ;length:0;number:0;"
- TestBatch14Mutants/is_image.a: batch14_test.go:135: compiled semantic mutant caught at line 2426: got "false/segment:,;" Go "false/segment:,;url:0;"
- TestBatch14Mutants/is_image.a#01: batch14_test.go:135: compiled semantic mutant caught at line 7484: got "true/segment:,;" Go "false/segment:,;"
- TestBatch14Mutants/is_image.a#02: batch14_test.go:135: compiled semantic mutant caught at line 5321: got "false/segment:,;url:0;" Go "true/segment:,;url:0;"
- TestBatch14Mutants/is_image.a#03: batch14_test.go:135: compiled semantic mutant caught at line 5288: got "false/segment:,;url:0;" Go "true/segment:,;url:0;"
- TestBatch14Mutants/is_image.a#04: batch14_test.go:135: compiled semantic mutant caught at line 3674: got "false/segment:,;url:0;" Go "true/segment:,;url:0;"
- TestBatch14Mutants/is_image.a#05: batch14_test.go:135: compiled semantic mutant caught at line 2099: got "true/segment:,;url:0;" Go "false/segment:,;url:0;"
- TestBatch14Mutants/is_image.a#06: batch14_test.go:135: compiled semantic mutant caught at line 2: got "false/segment:,;url:-1;" Go "false/segment:,;url:0;"
- TestBatch14Mutants/is_image.a#07: batch14_test.go:135: compiled semantic mutant caught at line 2: got "false/segment:,;wrong-value;url:0;" Go "false/segment:,;url:0;"
- TestBatch14Mutants/is_image.a#08: batch14_test.go:135: compiled semantic mutant caught at line 113: got "false/segment:,;url:0;url:1;" Go "false/segment:,;url:0;"
- TestBatch14Mutants/is_image.a#09: batch14_test.go:135: compiled semantic mutant caught at line 6473: got "false/segment:,;url:0;" Go "true/segment:,;url:0;"
- TestBatch14Mutants/is_background_position.a: batch14_test.go:135: compiled semantic mutant caught at line 1308: got "false/segment: ;length:0;percentage:0;length:1;percentage:1;" Go "true/segment: ;length:0;percentage:0;"
- TestBatch14Mutants/is_background_position.a#01: batch14_test.go:135: compiled semantic mutant caught at line 1365: got "false/segment: ;length:0;percentage:0;length:1;percentage:1;" Go "true/segment: ;length:0;percentage:0;"
- TestBatch14Mutants/is_background_position.a#02: batch14_test.go:135: compiled semantic mutant caught at line 7485: got "true/segment: ;" Go "false/segment: ;"
- TestBatch14Mutants/is_background_position.a#03: batch14_test.go:135: compiled semantic mutant caught at line 3: got "false/segment: ;length:0;" Go "false/segment: ;length:0;percentage:0;"
- TestBatch14Mutants/is_background_position.a#04: batch14_test.go:135: compiled semantic mutant caught at line 3: got "false/segment: ;percentage:0;length:0;percentage:0;" Go "false/segment: ;length:0;percentage:0;"
- TestBatch14Mutants/is_background_position.a#05: batch14_test.go:135: compiled semantic mutant caught at line 3: got "false/segment: ;length:-1;percentage:0;" Go "false/segment: ;length:0;percentage:0;"
- TestBatch14Mutants/is_background_position.a#06: batch14_test.go:135: compiled semantic mutant caught at line 3: got "false/segment:,;length:0;percentage:0;" Go "false/segment: ;length:0;percentage:0;"
- TestBatch15Mutants/push_jump.a: batch15_test.go:126: compiled semantic mutant caught at line 90: got "1/2:1:0:false:false:/labels:0;" Go "1/2:1:0:true:false:/labels:0;"
- TestBatch15Mutants/push_jump.a#01: batch15_test.go:126: compiled semantic mutant caught at line 90: got "1/2:1:0:true:true:/labels:0;" Go "1/2:1:0:true:false:/labels:0;"
- TestBatch15Mutants/push_jump.a#02: batch15_test.go:126: compiled semantic mutant caught at line 90: got "1/2:1:0:true:false:/labels:1;" Go "1/2:1:0:true:false:/labels:0;"
- TestBatch15Mutants/push_jump.a#03: batch15_test.go:126: compiled semantic mutant caught at line 90: got "1/1:2:0:true:false:/labels:0;" Go "1/2:1:0:true:false:/labels:0;"
- TestBatch15Mutants/push_jump.a#04: batch15_test.go:126: compiled semantic mutant caught at line 90: got "1/2:1:-1:true:false:/labels:0;" Go "1/2:1:0:true:false:/labels:0;"
- TestBatch15Mutants/pop_jump.a: batch15_test.go:126: compiled semantic mutant caught at line 125: got "0/" Go "1/2:1:6:true:false:/"
- TestBatch15Mutants/pop_jump.a#01: batch15_test.go:126: compiled semantic mutant caught at line 92: got "1/2:1:0:true:false:/" Go "0/"
- TestBatch15Mutants/make_unreachable.a: batch15_test.go:126: compiled semantic mutant caught at line 6770: got "0:false:false/1,/true,|1:false:false//nil" Go "0:false:false/1,/false,|1:false:false//nil"
- TestBatch15Mutants/make_unreachable.a#01: batch15_test.go:126: compiled semantic mutant caught at line 1: got "2:true:true/2,/nil|3:false:false//nil" Go "2:true:true/3,/nil|3:false:false//nil"
- TestBatch15Mutants/make_unreachable.a#02: batch15_test.go:126: compiled semantic mutant caught at line 1: got "2:true:true/3,/nil|2:true:true/3,/nil" Go "2:true:true/3,/nil|3:false:false//nil"
- TestBatch15Mutants/make_unreachable.a#03: batch15_test.go:126: compiled semantic mutant caught at line 1: got "2:true:true/3,/nil|3:true:false//nil" Go "2:true:true/3,/nil|3:false:false//nil"
- TestBatch15EmptyStackRefusal: batch15_test.go:189: compiled empty-stack guard mutant caught: baseline exit 70, mutant exit 0 and survived
- TestBatch2Mutants/intrinsic_element_named.a: batch2_test.go:101: compiled semantic mutant caught at line 24: got "true" Go "false"
- TestBatch2Mutants/hole_edges.a: batch2_test.go:101: compiled semantic mutant caught at line 179230: got "false,false" Go "true,false"
- TestBatch2Mutants/read_class_values.a: batch2_test.go:101: compiled semantic mutant caught at line 185959: got "Attribute::" Go "Callee::"
- TestBatch3Mutants/hex_value.a: batch3_test.go:99: compiled semantic mutant caught at line 43: got "-1" Go "15"
- TestBatch3Mutants/unescape_string_literal_text.a: batch3_test.go:99: compiled semantic mutant caught at line 1114212: got "0,99,111,112,121,10,:99,111,112,121," Go "0,169,10,:99,111,112,121,"
- TestBatch3Mutants/parameter_nodes.a: batch3_test.go:99: compiled semantic mutant caught at line 1115385: got "2,-1,3,2" Go "-1,-1,3,2"
- TestBatch3Mutants/parameter_nodes.a#01: batch3_test.go:99: compiled semantic mutant caught at line 1115375: got "0,1" Go ""
- TestBatch4Mutants/escape_terminator.a: batch4_test.go:85: compiled semantic mutant caught at line 13: got "false" Go "true"
- TestBatch4Mutants/followed_by_whitespace.a: batch4_test.go:85: compiled semantic mutant caught at line 3585: got "true:10" Go "false:10,11"
- TestBatch4Mutants/ignored_theme_key.a: batch4_test.go:85: compiled semantic mutant caught at line 75243: got "true" Go "false"
- TestBatch5Mutants/split_theme_key.a: batch5_test.go:92: compiled semantic mutant caught at line 283: got "1:" Go "0:"
- TestBatch5Mutants/split_theme_key.a#01: batch5_test.go:92: compiled semantic mutant caught at line 286: got "0:" Go "1:"
- TestBatch5Mutants/split_theme_key.a#02: batch5_test.go:92: compiled semantic mutant caught at line 288: got "1:99,104,97,110,103,101,100," Go "1:"
- TestBatch5Mutants/join_segments.a: batch5_test.go:92: compiled semantic mutant caught at line 290: got "0," Go "0,45,"
- TestBatch5Mutants/breakpoint_group_order.a: batch5_test.go:92: compiled semantic mutant caught at line 9639: got "2:true" Go "0:false"
- TestBatch5Mutants/breakpoint_group_order.a#01: batch5_test.go:92: compiled semantic mutant caught at line 9641: got "23:true" Go "-17:true"
- TestBatch6Mutants/at_rule.a: batch6_test.go:85: compiled semantic mutant caught at line 1: got "rule|||112,97,114,97,109,115,58,|||false|false|false|false|0|0|" Go "at-rule|||112,97,114,97,109,115,58,|||false|false|false|false|0|0|"
- TestBatch6Mutants/at_rule.a#01: batch6_test.go:85: compiled semantic mutant caught at line 14: got "1:1:true:0" Go "1:1:true:-1"
- TestBatch6Mutants/at_rule.a#02: batch6_test.go:85: compiled semantic mutant caught at line 14: got "0:0:true:" Go "1:1:true:-1"
- TestBatch6Mutants/at_rule.a#03: batch6_test.go:85: compiled semantic mutant caught at line 3: got "at-rule|99,104,97,110,103,101,100,|99,104,97,110,103,101,100,|112,97,114,97,109,115,58,|||false|false|false|false|0|0|" Go "at-rule|||112,97,114,97,109,115,58,|||false|false|false|false|0|0|"
- TestBatch6Mutants/style_rule.a: batch6_test.go:85: compiled semantic mutant caught at line 4: got "at-rule||||||false|false|false|false|0|0|" Go "rule||||||false|false|false|false|0|0|"
- TestBatch6Mutants/style_rule.a#01: batch6_test.go:85: compiled semantic mutant caught at line 17: got "1:1:true:0" Go "1:1:true:-1"
- TestBatch6Mutants/style_rule.a#02: batch6_test.go:85: compiled semantic mutant caught at line 17: got "0:0:true:" Go "1:1:true:-1"
- TestBatch6Mutants/style_rule.a#03: batch6_test.go:85: compiled semantic mutant caught at line 6: got "rule|99,104,97,110,103,101,100,|99,104,97,110,103,101,100,||||false|false|false|false|0|0|" Go "rule||||||false|false|false|false|0|0|"
- TestBatch6Mutants/variant_next_order.a: batch6_test.go:85: compiled semantic mutant caught at line 29488: got "-9007199254740989" Go "-9007199254740991"
- TestBatch6Mutants/variant_next_order.a#01: batch6_test.go:85: compiled semantic mutant caught at line 29485: got "-9007199254740988" Go "-9007199254740989"
- TestBatch6Mutants/variant_next_order.a#02: batch6_test.go:85: compiled semantic mutant caught at line 29486: got "false:0:-9007199254740989" Go "false:0:-9007199254740990"
- TestBatch7Mutants/new_variant_registry.a: batch7_test.go:85: compiled semantic mutant caught at line 1: got "1:false:0:0:0" Go "0:false:0:0:0"
- TestBatch7Mutants/new_variant_registry.a#01: batch7_test.go:85: compiled semantic mutant caught at line 25: got "84:false:0:2:2" Go "0:false:0:0:0"
- TestBatch7Mutants/register.a: batch7_test.go:85: compiled semantic mutant caught at line 8: got ":84:replacement" Go ":83:replacement"
- TestBatch7Mutants/register.a#01: batch7_test.go:85: compiled semantic mutant caught at line 680: got ":83:static" Go ":-17:static"
- TestBatch7Mutants/register.a#02: batch7_test.go:85: compiled semantic mutant caught at line 679: got "-17:true:-17:1:0" Go "82:true:-17:1:0"
- TestBatch7Mutants/attach_comparison.a: batch7_test.go:85: compiled semantic mutant caught at line 16: got "84:false:0:2:0" Go "84:false:0:2:1"
- TestBatch7Mutants/attach_comparison.a#01: batch7_test.go:85: compiled semantic mutant caught at line 21: got "-8" Go "15"
- TestBatch7Mutants/attach_comparison.a#02: batch7_test.go:85: compiled semantic mutant caught at line 15: got "missing" Go "-8"
- TestBatch8Mutants/recursively_decode_arbitrary_values.a: batch8_test.go:85: compiled semantic mutant caught at line 413: got "function:102,110,|function:118,97,114,|word:45,45,97,95,98,|separator:44,|word:32,99,32,100,|separator:44,|function:117,114,108,|word:97,32,98,|" Go "function:102,110,|function:118,97,114,|word:45,45,97,95,98,|separator:44,|word:32,99,32,100,|separator:44,|function:117,114,108,|word:97,95,98,|"
- TestBatch8Mutants/recursively_decode_arbitrary_values.a#01: batch8_test.go:85: compiled semantic mutant caught at line 71: got "function:99,97,108,99,|function:118,97,114,|word:45,45,97,32,98,|word:43,49,112,120,|" Go "function:99,97,108,99,|function:118,97,114,|word:45,45,97,95,98,|word:43,49,112,120,|"
- TestBatch8Mutants/recursively_decode_arbitrary_values.a#02: batch8_test.go:85: compiled semantic mutant caught at line 714: got "unknown:97,32,117,114,108,|word:45,45,97,95,98,|word:99,92,95,100,95,101,|word:95,102,95,103,|function:102,111,111,95,98,97,114,|word:104,95,105,|" Go "unknown:97,95,117,114,108,|word:45,45,97,95,98,|word:99,92,95,100,95,101,|word:95,102,95,103,|function:102,111,111,95,98,97,114,|word:104,95,105,|"
- TestBatch8Mutants/decode_arbitrary_value.a: batch8_test.go:85: compiled semantic mutant caught at line 70: got "99,97,108,99,40,49,112,120,43,50,112,120,41," Go "99,97,108,99,40,49,112,120,32,43,32,50,112,120,41,"
- TestBatch8Mutants/decode_arbitrary_value.a#01: batch8_test.go:85: compiled semantic mutant caught at line 70: got "99,97,108,99,40,49,112,120,43,50,112,120,41," Go "99,97,108,99,40,49,112,120,32,43,32,50,112,120,41,"
- TestBatch8Mutants/decode_arbitrary_value.a#02: batch8_test.go:85: compiled semantic mutant caught at line 30: got "117,110,101,120,112,101,99,116,101,100,32,109,97,116,104,32,105,110,112,117,116,58,117,110,101,120,112,101,99,116,101,100,32,112,97,114,115,101,32,105,110,112,117,116,58,85,82,76,40,97,32,98,41,32," Go "85,82,76,40,97,32,98,41,"
- TestBatch8Mutants/register_theme_breakpoint_variants.a: batch8_test.go:85: compiled semantic mutant caught at line 851: got "49,48,48,:-17:static|101,120,105,115,116,105,110,103,:1:functional|115,109,:2:compound|" Go "101,120,105,115,116,105,110,103,:1:functional|115,109,:2:compound|"
- TestBatch8Mutants/register_theme_breakpoint_variants.a#01: batch8_test.go:85: compiled semantic mutant caught at line 855: got "49,48,48,:-17:static|101,120,105,115,116,105,110,103,:1:functional|115,109,:2:compound|" Go "101,120,105,115,116,105,110,103,:1:functional|115,109,:2:compound|"
- TestBatch8Mutants/register_theme_breakpoint_variants.a#02: batch8_test.go:85: compiled semantic mutant caught at line 747: got "101,120,105,115,116,105,110,103,:1:static|115,109,:2:compound|" Go "101,120,105,115,116,105,110,103,:1:functional|115,109,:2:compound|"
- TestBatch8Mutants/register_theme_breakpoint_variants.a#03: batch8_test.go:85: compiled semantic mutant caught at line 859: got "49,48,48,:-16:static|101,120,105,115,116,105,110,103,:1:functional|115,109,:2:compound|" Go "49,48,48,:-17:static|101,120,105,115,116,105,110,103,:1:functional|115,109,:2:compound|"
- TestBatch9Mutants/parse_css.a: batch9_test.go:115: compiled semantic mutant caught at line 31: got "ok:[rule:46,97,:::::false:false:true:[declaration::::99,111,108,111,114,:114,101,100,:false:true:false:[]|]|]" Go "ok:[comment:::::33,32,108,105,99,101,110,115,101,32,:false:false:false:[]|rule:46,97,:::::false:false:true:[declaration::::99,111,108,111,114,:114,101,100,:false:true:false:[]|]|]"
- TestBatch9Mutants/parse_css.a#01: batch9_test.go:115: compiled semantic mutant caught at line 418: got "ok:[rule:117,110,101,120,112,101,99,116,101,100,32,116,114,105,109,58,65279,46,97,32,:::::false:false:true:[declaration::::99,111,108,111,114,:114,101,100,:false:true:false:[]|]|]" Go "ok:[rule:46,97,:::::false:false:true:[declaration::::99,111,108,111,114,:114,101,100,:false:true:false:[]|]|]"
- TestBatch9Mutants/parse_css.a#02: batch9_test.go:115: compiled semantic mutant caught at line 7: got "ok:[]" Go "ok:[declaration::::45,45,120,::false:true:false:[]|]"
- TestBatch9Mutants/parse_css.a#03: batch9_test.go:115: compiled semantic mutant caught at line 21: got "ok:[rule:46,97,:::::false:false:true:[unexpected declaration:::::98,97,99,107,103,114,111,117,110,100,58,32,117,114,108,40,97,:false:false:false:[]|unexpected declaration:::::98,46,112,110,103,41,:false:false:false:[]|]|]" Go "ok:[rule:46,97,:::::false:false:true:[declaration::::98,97,99,107,103,114,111,117,110,100,:117,114,108,40,97,59,98,46,112,110,103,41,:false:true:false:[]|]|]"
- TestBatch9Mutants/parse_css.a#04: batch9_test.go:115: compiled semantic mutant caught at line 27: got "ok:[]" Go "error:5:77,105,115,115,105,110,103,32,99,108,111,115,105,110,103,32,125,32,97,116,32,46,97,"
- TestBatch9Mutants/parse_css.a#05: batch9_test.go:115: compiled semantic mutant caught at line 5: got "error:1:77,105,115,115,105,110,103,32,111,112,101,110,105,110,103,32,40," Go "error:0:77,105,115,115,105,110,103,32,111,112,101,110,105,110,103,32,40,"
- TestBatch9Mutants/design_system_decline_message.a: batch9_test.go:115: compiled semantic mutant caught at line 433: got "32,99,111,117,108,100,32,110,111,116,32,98,117,105,108,100,32,116,104,105,115,32,112,114,111,106,101,99,116,39,115,32,84,97,105,108,119,105,110,100,32,100,101,115,105,103,110,32,115,121,115,116,101,109,44,32,115,111,32,105,116,32,105,115,32,114,101,112,111,114,116,105,110,103,32,110,111,116,104,105,110,103,32,114,97,116,104,101,114,32,116,104,97,110,32,114,101,112,111,114,116,105,110,103,32,97,32,99,108,101,97,110,32,116,114,101,101,58,32,60,110,105,108,62," Go "32,99,111,117,108,100,32,110,111,116,32,98,117,105,108,100,32,116,104,105,115,32,112,114,111,106,101,99,116,39,115,32,84,97,105,108,119,105,110,100,32,100,101,115,105,103,110,32,115,121,115,116,101,109,32,102,114,111,109,32,116,104,101,109,101,46,99,115,115,44,32,115,111,32,105,116,32,105,115,32,114,101,112,111,114,116,105,110,103,32,110,111,116,104,105,110,103,32,114,97,116,104,101,114,32,116,104,97,110,32,114,101,112,111,114,116,105,110,103,32,97,32,99,108,101,97,110,32,116,114,101,101,58,32,60,110,105,108,62,"
- TestBatch9Mutants/design_system_decline_message.a#01: batch9_test.go:115: compiled semantic mutant caught at line 421: got "32,99,111,117,108,100,32,110,111,116,32,98,117,105,108,100,32,116,104,105,115,32,112,114,111,106,101,99,116,39,115,32,84,97,105,108,119,105,110,100,32,100,101,115,105,103,110,32,115,121,115,116,101,109,44,32,115,111,32,105,116,32,105,115,32,114,101,112,111,114,116,105,110,103,32,110,111,116,104,105,110,103,32,114,97,116,104,101,114,32,116,104,97,110,32,114,101,112,111,114,116,105,110,103,32,97,32,99,108,101,97,110,32,116,114,101,101,58,32," Go "32,99,111,117,108,100,32,110,111,116,32,98,117,105,108,100,32,116,104,105,115,32,112,114,111,106,101,99,116,39,115,32,84,97,105,108,119,105,110,100,32,100,101,115,105,103,110,32,115,121,115,116,101,109,44,32,115,111,32,105,116,32,105,115,32,114,101,112,111,114,116,105,110,103,32,110,111,116,104,105,110,103,32,114,97,116,104,101,114,32,116,104,97,110,32,114,101,112,111,114,116,105,110,103,32,97,32,99,108,101,97,110,32,116,114,101,101,58,32,60,110,105,108,62,"
- TestBatch9Mutants/loaded_theme.a: batch9_test.go:115: compiled semantic mutant caught at line 734: got "-1:false:true" Go "0:true:true"
- TestBatch9Mutants/loaded_theme.a#01: batch9_test.go:115: compiled semantic mutant caught at line 734: got "0:true:false" Go "0:true:true"
- TestSlot03HelperMutants/component_base_name.a/value: slot03_test.go:160: compiled semantic mutant caught at line 159: got "false" Go "true"
- TestSlot03HelperMutants/tailwind_space.a/value: slot03_test.go:160: compiled semantic mutant caught at line 847: got "false" Go "true"
- TestSlot03HelperMutants/listener_kinds.a/value: slot03_test.go:160: compiled semantic mutant caught at line 1114948: got "JsxAttribute,CallExpression,StringLiteral" Go "JsxAttribute,CallExpression,VariableDeclaration"
- TestSlot03HelperMutants/listener_kinds.a/shared-list: slot03_test.go:160: compiled semantic mutant caught at line 1114949: got "StringLiteral,CallExpression,VariableDeclaration" Go "JsxAttribute,CallExpression,VariableDeclaration"
- TestConsumerCoverageRejectsMutant: slot03_test.go:277: missing-consumer mutant caught: consumer capture mismatch: missing [better-tailwindcss/no-unknown-classes] extra []
