Built isLineWidth, isImage and isBackgroundPosition in separate .a files: twelve dependency edges across four rules.
Commits: claim 9c08c8b6 pushed before source; base c01907a; implementation bb420a1b2e5f4ad78df3dedfe1295c98fdc2f4b6 published on codex/lint-helpers-03 only.
Commands: focused gate PASS 58.003s, 8,610 verdict/trace lines; full owned gate PASS 458.164s, 3,034,762 comparison lines; shared helpers PASS 71.045s; vet/format clean; six uncached input probes PASS 1.733s.
Mutants: twenty-three new compiling variants caught; all one hundred thirty-four owned variants, four inherited variants and missing-consumer check caught.
Not covered: full repository gate, native dependency wiring, whole-rule findings/fixes/suggestions or unavailable live Tailwind/corpora.

## Landing and claims

Only codex/lint-helpers-03 is pushed by this worker. Earlier thirty-nine helpers are complete, oracle-green on c01907a and pushed at 02369a49 before new selection. Fetched main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06 is already an ancestor; no rebase is needed. Every claims file on all twenty origin/codex/lint-helpers* branches is inspected. Shared HELPERS.md reserves the comments bundle; custom-option decoding is rule-local. The three selected concrete helpers tie the highest available consumer count at four. Claim 9c08c8b6 is pushed before source. A post-claim fetch confirms each appears only in slot 03's claims file. No fourth helper is claimed.

The user-named harness commit ab70f38d4 is not on this fetched main. This helper unit requires no Diagnostic or context.report changes. No shared harness, generator, compiler, registration or listener routing is edited. Allocator-aware shared test changes will be accepted through main when they arrive; none is reverted.

## Behavior, observations and boundaries

isLineWidth obtains space segments and calls length then number with short-circuit evaluation. Failing those, it accepts only thin, medium or thick. Any other part returns false; finishing the loop returns true, including an empty segment list supplied by a dependency.

isImage obtains comma segments. Exact lowercase var( parts are skipped without counting. URLs are checked first, then six conic/linear/radial gradient prefixes with optional repeating-, then element/image/cross-fade/image-set prefixes. Any other part rejects the whole string. At least one non-variable accepted part must count. Prefix tests deliberately accept malformed/unclosed prefixes just as the Go implementation does; case and leading characters matter.

isBackgroundPosition obtains space segments. Exact center/top/right/bottom/left keywords count before predicates; var( parts skip without counting. Length then percentage use lazy OR. Any other part rejects; at least one non-variable accepted part must count.

Prefix checks are JS RegExp literals with the u flag. The actual Go source uses HasPrefix/list comparisons and documents the original image regexes; the shared regex table at 071fb012848ce0408428c61aba0857cca472236f has 107 compile sites and no data_type.go row. Exact keyword sets use Array.includes. There is no handwritten regex engine, runtime option pattern or finding position in this batch.

The temporary owned Go overlay leaves the three helper bodies unchanged and renames dependency definitions only to install recording wrappers. Wrappers call real Go segmentation and real length/number/percentage/URL predicates. They record separators, unchanged input and call order. Expected results therefore come from actual Go helper control flow. Test-only Adamic dependencies replay real Go predicate observations and parts, recording the same contracts. Duplicate parts share the first equal-part index in both traces because predicates are pure. This verifies these helper compositions, not the separately owned native dependency implementations. Consumers must wire actual dependencies before whole-rule integration. Malformed Unicode byte strings and unpaired UTF-16 are outside the adapter boundary.

All four consumer rules supply 118 runtime sources. Full sources, extracted tokens, numeric/unit/math controls, every space/comma pair of 32 parts, and ten image-function prefix families produce 2,870 unique strings. Each exercises all three helpers: 8,610 verdict/trace lines. Actual Go, Node source, emitted JavaScript and ASan/UBSan native with Linux leak checks agree byte for byte. Source/emitted Node really execute the literal regexes.

## Consumers and remaining blockers

Each of the three helpers removes one dependency from each of:

- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

That is twelve removed edges, four distinct consumers and zero additional fully helper-ready rules. readiness.json subtracts only this worker's forty-two delivered helpers from the base ledger; other workers' claims are not treated as available code.

## Commands and evidence

Setup: `bash cloud/setup.sh`, then `source /workspace/adamic-tools/env.sh`. Go ready 0s; clang ready 1s; Node ready 1s; submodules ready 1s; build cache warm 46s; done 46s. nproc=5, cgroup cpu.max=400000 100000, memory=17.6GB. Toolchain Go1.27.1, clang20.1.8, Node24.19.0.

- `go test ./stage1/cohere/lint/helpers/slot03 -run TestBatch14 -count=1 -v -timeout=20m` -> evidence/focused.log, PASS 58.003s.
- `go test ./stage1/cohere/lint/helpers ./stage1/cohere/lint/helpers/slot03 -count=1 -v -timeout=20m` -> evidence/final.log.
- `go vet ./...` -> evidence/vet.log, exit 0, no output.
- `gofmt -l stage1/cohere/lint/helpers/slot03` -> evidence/format.log, no output.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m` -> evidence/input-oracle.log, six probes PASS 1.733s, zero cache hits.
- `python3 stage1/cohere/lint/helpers/slot03/batch14/testdata/regenerate.py` -> evidence/regenerate.log and regenerate-repeat.log; both capture 118 sources and reproduce identical sources.jsonl.gz and coverage.json SHA256 hashes in evidence/reproduce.log.

Capture runs the upstream Go Tailwind package with a temporary observation overlay. Its exit 1 is explicitly retained: eight known external-engine/corpus guards fail, including the unavailable /Users/kirkouimet/Projects/ahra/app/_theme/styles path. The source capture remains complete for all four selected consumers; this is not a passing whole-rule package gate. No shared file is changed to suppress the guards.

## Every new mutant and its catch

All mutant copies compile and execute with exit 0 and empty stderr before changed stdout counts. Width mutations remove thick, eagerly call number, corrupt length input, corrupt segment input, change separator and accept invalid parts. Image mutations change variable case, accept zero count, drop repeating/radial/cross-fade prefixes, change image case, corrupt URL/segment arguments, skip invalid parts and omit URL count. Position mutations remove bottom, disable variable skipping, accept zero count, replace OR with AND, eagerly call percentage, corrupt length input and change separator.

- TestBatch14Mutants/is_line_width.a: 955: got "false/segment: ;length:0;number:0;length:1;number:1;" Go "true/segment: ;length:0;number:0;length:1;number:1;"
- TestBatch14Mutants/is_line_width.a#01: 1: got "false/segment: ;number:0;length:0;number:0;" Go "false/segment: ;length:0;number:0;"
- TestBatch14Mutants/is_line_width.a#02: 1: got "false/segment: ;length:-1;number:0;" Go "false/segment: ;length:0;number:0;"
- TestBatch14Mutants/is_line_width.a#03: 1: got "false/segment: ;wrong-value;length:0;number:0;" Go "false/segment: ;length:0;number:0;"
- TestBatch14Mutants/is_line_width.a#04: 1: got "false/segment:,;length:0;number:0;" Go "false/segment: ;length:0;number:0;"
- TestBatch14Mutants/is_line_width.a#05: 1: got "true/segment: ;length:0;number:0;" Go "false/segment: ;length:0;number:0;"
- TestBatch14Mutants/is_image.a: 2426: got "false/segment:,;" Go "false/segment:,;url:0;"
- TestBatch14Mutants/is_image.a#01: 7484: got "true/segment:,;" Go "false/segment:,;"
- TestBatch14Mutants/is_image.a#02: 5321: got "false/segment:,;url:0;" Go "true/segment:,;url:0;"
- TestBatch14Mutants/is_image.a#03: 5288: got "false/segment:,;url:0;" Go "true/segment:,;url:0;"
- TestBatch14Mutants/is_image.a#04: 3674: got "false/segment:,;url:0;" Go "true/segment:,;url:0;"
- TestBatch14Mutants/is_image.a#05: 2099: got "true/segment:,;url:0;" Go "false/segment:,;url:0;"
- TestBatch14Mutants/is_image.a#06: 2: got "false/segment:,;url:-1;" Go "false/segment:,;url:0;"
- TestBatch14Mutants/is_image.a#07: 2: got "false/segment:,;wrong-value;url:0;" Go "false/segment:,;url:0;"
- TestBatch14Mutants/is_image.a#08: 113: got "false/segment:,;url:0;url:1;" Go "false/segment:,;url:0;"
- TestBatch14Mutants/is_image.a#09: 6473: got "false/segment:,;url:0;" Go "true/segment:,;url:0;"
- TestBatch14Mutants/is_background_position.a: 1308: got "false/segment: ;length:0;percentage:0;length:1;percentage:1;" Go "true/segment: ;length:0;percentage:0;"
- TestBatch14Mutants/is_background_position.a#01: 1365: got "false/segment: ;length:0;percentage:0;length:1;percentage:1;" Go "true/segment: ;length:0;percentage:0;"
- TestBatch14Mutants/is_background_position.a#02: 7485: got "true/segment: ;" Go "false/segment: ;"
- TestBatch14Mutants/is_background_position.a#03: 3: got "false/segment: ;length:0;" Go "false/segment: ;length:0;percentage:0;"
- TestBatch14Mutants/is_background_position.a#04: 3: got "false/segment: ;percentage:0;length:0;percentage:0;" Go "false/segment: ;length:0;percentage:0;"
- TestBatch14Mutants/is_background_position.a#05: 3: got "false/segment: ;length:-1;percentage:0;" Go "false/segment: ;length:0;percentage:0;"
- TestBatch14Mutants/is_background_position.a#06: 3: got "false/segment:,;length:0;percentage:0;" Go "false/segment: ;length:0;percentage:0;"

The final regression log records each earlier semantic mutant and inherited missing-consumer coverage check as well. Full repository testing, arbitrary dependency implementations and integrated native utility-engine behavior remain outside this unit.

## Final regression mutant witnesses

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
