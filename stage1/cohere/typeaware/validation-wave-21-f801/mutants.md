# Current-main mutant ledger

Fresh output from the f8013f0ba rebase. Native rule mutants compile and exit successfully; independent byte comparisons catch them. Prerequisite, compiler-question, framing and refusal mutants have their separately stated expectations.

- wave_21_core_test.go:94: obj-calls judgment/edit mutant: exit 0; byte comparison catches byte 552
- wave_21_core_test.go:94: object-constructor judgment/edit mutant: exit 0; byte comparison catches byte 48057
- wave_21_core_test.go:94: promise-return judgment/edit mutant: exit 0; byte comparison catches byte 80630
- wave_21_core_test.go:107: reference-node: released panic 70; retaining registry mutant exits 0 and is caught
- wave_21_core_test.go:123: reference suffix mutant compiles; malformed-question assertion catches it
- wave_21_listeners_test.go:157: 13 numeric listener declarations match production Go, normal and ASAN; wrong-listener mutant exits 0 with empty stderr, byte comparison catches byte 15
- wave_21_next_test.go:272: collection mutant: exit 0, comparison catches byte 1501
- wave_21_next_test.go:272: outcome mutant: exit 0, comparison catches byte 7963
- wave_21_next_test.go:272: pure mutant: exit 0, comparison catches byte 18824
- wave_21_next_test.go:299: awaited-shape: released panic 70; retaining registry mutant exits 0 and is caught
- wave_21_next_test.go:299: type-declaration-ancestors: released panic 70; retaining registry mutant exits 0 and is caught
- wave_21_process_test.go:121: timer judgment mutant: exit 0, comparison catches byte 102
- wave_21_process_test.go:121: process-exit judgment mutant: exit 0, comparison catches byte 6479
- wave_21_process_test.go:121: blocking judgment mutant: exit 0, comparison catches byte 15965
- wave_21_process_test.go:136: node-symbol-context: released panic 70; retaining registry mutant exits 0 and is caught
- wave_21_process_test.go:136: node-symbol-context/follow-alias: released panic 70; retaining registry mutant exits 0 and is caught
- wave_21_process_test.go:136: program-modules: released panic 70; retaining registry mutant exits 0 and is caught
- wave_21_process_test.go:136: resolved-declaration: released panic 70; retaining registry mutant exits 0 and is caught
- wave_21_react_cores_test.go:218: adjacency-row alias mutant: exit 0, empty stderr; byte comparison catches byte 3251
- wave_21_react_cores_test.go:257: released handle exits 70; retained registry mutant exits 0 and is caught
- wave_21_react_cores_test.go:263: react-type-names suffix mutant compiles; strict-request assertion catches it
- wave_21_react_prerequisites_test.go:114: effect-positive prerequisite guard mutant: exit 0; required Go message disappeared
- wave_21_react_prerequisites_test.go:114: render-positive prerequisite guard mutant: exit 0; required Go message disappeared
- wave_21_react_prerequisites_test.go:114: static-positive prerequisite guard mutant: exit 0; required Go message disappeared
- wave_21_react_prerequisites_test.go:129: parser guard mutant: exit 0; required JSX refusal disappeared
- wave_21_test.go:89: enum mutant: exit 0, byte oracle catches byte 65
- wave_21_test.go:103: released registry mutant: exit 0 caught by required panic 70
- coverage_test.go:187: released-registry mutant exits 0, caught by the required panic 70
- typeaware_test.go:284: wrong-node type mutant: cohere byte oracle catches byte 480; findings 0 queries 61
- typeaware_test.go:301: released registry deletion mutant: stale-query expectation catches exit 0 instead of 70
- typeaware_test.go:317: empty-frame mutant: panic 70, checker returned no type parts
- typeaware_test.go:317: missing-header mutant: panic 70, invalid checker type frame
- typeaware_test.go:317: bad-length mutant: panic 70, invalid checker type length or flags
- typeaware_test.go:340: exact kind guard mutant: mismatch-refusal expectation catches exit 0
- volume_test.go:162: assignable-types mutant: byte oracle catches byte 81; findings 67
- volume_test.go:162: widened-shape mutant: byte oracle catches byte 562; findings 74
- volume_test.go:162: enum-types mutant: byte oracle catches byte 6039; findings 74
- volume_test.go:162: type-symbol mutant: byte oracle catches byte 21217; findings 73
- volume_test.go:162: scope-locals mutant: byte oracle catches byte 8652; findings 71
- volume_test.go:162: call-returns mutant: byte oracle catches byte 13238; findings 77
- volume_test.go:162: property-shape mutant: byte oracle catches byte 7557; findings 72
- volume_test.go:162: contextual-shape mutant: byte oracle catches byte 8118; findings 72
- volume_test.go:162: symbol-origin mutant: byte oracle catches byte 26042; findings 74
- volume_test.go:162: type-origin mutant: byte oracle catches byte 25376; findings 75
- volume_test.go:162: property-info mutant: byte oracle catches byte 21519; findings 63
- volume_test.go:162: call-count mutant: byte oracle catches byte 16475; findings 75
- volume_test.go:162: call-parameters mutant: byte oracle catches byte 16475; findings 75
- volume_test.go:162: apparent-shape mutant: byte oracle catches byte 16475; findings 75
- volume_test.go:162: base-shapes mutant: byte oracle catches byte 25376; findings 74
- volume_test.go:185: released registry mutant: exit 0 caught by required panic 70
- volume_test.go:260: strict-this compiler-option mutant: refusal expectation catches exit 0 instead of 70
- volume_test.go:266: implicit-this mutant also differs from independent Go at byte 133

The expanded external Node oracle also passed TestTheOracleCatchesOneByte; see wave21-f801-node.log.
