Unit u036: all 11 assigned internal/lower rows exist at e2492670b06a4dce1deafe837158ce0366cf2bc4.
Clean whole-package baseline: PASS, 34.628 seconds; nproc 5; no assigned-row skips.
Bounded verdicts: six sacred, three subsumed, two overlapping.
Two vacuous rows; 20 production mutants, two entry probes, five slice survivors.
Evidence: test-audit/internal-lower-library_object, review/test-audit/internal-lower-library_object/.

```json
[
  {
    "test": "TestObjectRefusalsExplainSoundness",
    "package": "internal/lower",
    "file": "internal/lower/library_object_test.go",
    "seconds": 0.293,
    "oracle": "Requires Refused plus handwritten reason substrings. Does not assert the whole diagnostic.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M02",
      "M04",
      "M13"
    ],
    "unique_kills": [
      "M01",
      "M02",
      "M04"
    ],
    "last_proven_fail": "M13: library_object_test.go:32: got /tmp/adamic-gate/TestObjectRefusalsExplainSoundnessproperty_descriptorsObject.ge2165319558/001/main.a:1:1: Adamic 0.1 refuses a method read as a value (getOwnPropertyDescriptors would lose its object, and this with it); call it in an arrow that keeps the object: (o) => Object.getOwnPropertyDescriptors(o) (unbound-method), want refusal with \"property descriptors\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": true,
    "matrix_rows": [
      "TestObjectRefusalsExplainSoundness",
      "TestObjectUnprovenShapesStayNotYet",
      "TestLibraryRegexOffsetRequiresNoCaptures",
      "TestLibraryStringRefusals",
      "TestConsoleLowersToWriteLine",
      "TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat",
      "TestWhatZeroOneRefusesIsRefusedWithAFix",
      "TestAMutableLocationSeenWiderIsRefused",
      "TestAViewThatCantWriteIsNotRefused",
      "TestATupleSeenAsAnArrayIsNotYet",
      "TestAMethodReadAsAValueIsRefused"
    ],
    "evidence": "ADAMIC_MUTANT=M13 ADAMIC_BUILD_CACHE_DIR=/tmp/u036/cache/M13 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; library_object_test.go:32: got /tmp/adamic-gate/TestObjectRefusalsExplainSoundnessproperty_descriptorsObject.ge2165319558/001/main.a:1:1: Adamic 0.1 refuses a method read as a value (getOwnPropertyDescriptors would lose its object, and this with it); call it in an arrow that keeps the object: (o) => Object.getOwnPropertyDescriptors(o) (unbound-method), want refusal with \"property descriptors\"",
    "subsumption_mutants": null
  },
  {
    "test": "TestObjectUnprovenShapesStayNotYet",
    "package": "internal/lower",
    "file": "internal/lower/library_object_test.go",
    "seconds": 0.129,
    "oracle": "Requires only errors.As(NotYet); ignores location and reason. M05 passes despite a wrong Object.collectBy diagnostic; M02 passes several cases with a different NotYet reason.",
    "oracle_kind": "self",
    "kills": [
      "M13",
      "M19"
    ],
    "unique_kills": [
      "M19"
    ],
    "last_proven_fail": "M19: library_object_test.go:52: got <nil>, want NotYet",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": true,
    "matrix_rows": [
      "TestObjectRefusalsExplainSoundness",
      "TestObjectUnprovenShapesStayNotYet",
      "TestLibraryRegexOffsetRequiresNoCaptures",
      "TestLibraryStringRefusals",
      "TestConsoleLowersToWriteLine",
      "TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat",
      "TestWhatZeroOneRefusesIsRefusedWithAFix",
      "TestAMutableLocationSeenWiderIsRefused",
      "TestAViewThatCantWriteIsNotRefused",
      "TestATupleSeenAsAnArrayIsNotYet",
      "TestAMethodReadAsAValueIsRefused"
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/u036/cache/M19-corrected timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; library_object_test.go:52: got <nil>, want NotYet",
    "subsumption_mutants": null
  },
  {
    "test": "TestLibraryRegexOffsetRequiresNoCaptures",
    "package": "internal/lower",
    "file": "internal/lower/library_regex_callback_shape_test.go",
    "seconds": 0.122,
    "oracle": "Requires a producer be visited and the capture-free proof return false for four unproven inputs. No positive capture-free control; its false entry probe passes.",
    "oracle_kind": "self",
    "kills": [
      "M07",
      "M09"
    ],
    "unique_kills": [
      "M07",
      "M09"
    ],
    "last_proven_fail": "M09: library_regex_callback_shape_test.go:36: capture-free producer proof admitted an unproven pattern",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "vacuous_subcases": [],
    "bounded": true,
    "matrix_rows": [
      "TestObjectRefusalsExplainSoundness",
      "TestObjectUnprovenShapesStayNotYet",
      "TestLibraryRegexOffsetRequiresNoCaptures",
      "TestLibraryStringRefusals",
      "TestConsoleLowersToWriteLine",
      "TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat",
      "TestWhatZeroOneRefusesIsRefusedWithAFix",
      "TestAMutableLocationSeenWiderIsRefused",
      "TestAViewThatCantWriteIsNotRefused",
      "TestATupleSeenAsAnArrayIsNotYet",
      "TestAMethodReadAsAValueIsRefused"
    ],
    "evidence": "ADAMIC_MUTANT=M09 ADAMIC_BUILD_CACHE_DIR=/tmp/u036/cache/M09 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; library_regex_callback_shape_test.go:36: capture-free producer proof admitted an unproven pattern",
    "subsumption_mutants": null
  },
  {
    "test": "TestLibraryStringRefusals",
    "package": "internal/lower",
    "file": "internal/lower/library_string_test.go",
    "seconds": 0.21,
    "oracle": "Requires an error containing a handwritten reason; does not require a particular error class.",
    "oracle_kind": "self",
    "kills": [
      "M13",
      "M15",
      "M20"
    ],
    "unique_kills": [],
    "last_proven_fail": "M20: library_string_test.go:27: want refusal containing \"ToPrimitive is not lowered\", got /tmp/adamic-gate/TestLibraryStringRefusalsobject_conversion2623050405/001/main.a:1:1: stage 0 can't lower console.log with other than one string argument yet",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestWhatZeroOneRefusesIsRefusedWithAFix"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 0.483,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": true,
    "matrix_rows": [
      "TestObjectRefusalsExplainSoundness",
      "TestObjectUnprovenShapesStayNotYet",
      "TestLibraryRegexOffsetRequiresNoCaptures",
      "TestLibraryStringRefusals",
      "TestConsoleLowersToWriteLine",
      "TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat",
      "TestWhatZeroOneRefusesIsRefusedWithAFix",
      "TestAMutableLocationSeenWiderIsRefused",
      "TestAViewThatCantWriteIsNotRefused",
      "TestATupleSeenAsAnArrayIsNotYet",
      "TestAMethodReadAsAValueIsRefused"
    ],
    "evidence": "ADAMIC_MUTANT=M20 ADAMIC_BUILD_CACHE_DIR=/tmp/u036/cache/M20 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; library_string_test.go:27: want refusal containing \"ToPrimitive is not lowered\", got /tmp/adamic-gate/TestLibraryStringRefusalsobject_conversion2623050405/001/main.a:1:1: stage 0 can't lower console.log with other than one string argument yet",
    "subsumption_mutants": 3
  },
  {
    "test": "TestConsoleLowersToWriteLine",
    "package": "internal/lower",
    "file": "internal/lower/lower_test.go",
    "seconds": 0.044,
    "oracle": "Exact handwritten IR statements, streams, string deduplication and source filename.",
    "oracle_kind": "self",
    "kills": [
      "M13",
      "M14",
      "M20"
    ],
    "unique_kills": [
      "M14"
    ],
    "last_proven_fail": "M20: lower_test.go:32: /tmp/adamic-gate/TestConsoleLowersToWriteLine2735757914/001/main.a:1:1: stage 0 can't lower console.log with other than one string argument yet",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": true,
    "matrix_rows": [
      "TestObjectRefusalsExplainSoundness",
      "TestObjectUnprovenShapesStayNotYet",
      "TestLibraryRegexOffsetRequiresNoCaptures",
      "TestLibraryStringRefusals",
      "TestConsoleLowersToWriteLine",
      "TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat",
      "TestWhatZeroOneRefusesIsRefusedWithAFix",
      "TestAMutableLocationSeenWiderIsRefused",
      "TestAViewThatCantWriteIsNotRefused",
      "TestATupleSeenAsAnArrayIsNotYet",
      "TestAMethodReadAsAValueIsRefused"
    ],
    "evidence": "ADAMIC_MUTANT=M20 ADAMIC_BUILD_CACHE_DIR=/tmp/u036/cache/M20 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; lower_test.go:32: /tmp/adamic-gate/TestConsoleLowersToWriteLine2735757914/001/main.a:1:1: stage 0 can't lower console.log with other than one string argument yet",
    "subsumption_mutants": null
  },
  {
    "test": "TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat",
    "package": "internal/lower",
    "file": "internal/lower/lower_test.go",
    "seconds": 0.478,
    "oracle": "Requires NotYet and handwritten diagnostic suffixes containing position and construct.",
    "oracle_kind": "self",
    "kills": [
      "M10",
      "M11",
      "M13",
      "M16",
      "M18",
      "M20"
    ],
    "unique_kills": [
      "M18"
    ],
    "last_proven_fail": "M20: lower_test.go:95: got /tmp/adamic-gate/TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhatan_arrow_r3582234619/001/main.a:3:2: stage 0 can't lower console.log with other than one string argument yet, want an error ending \"main.a:3:25: stage 0 can't lower a function returning never yet\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": true,
    "matrix_rows": [
      "TestObjectRefusalsExplainSoundness",
      "TestObjectUnprovenShapesStayNotYet",
      "TestLibraryRegexOffsetRequiresNoCaptures",
      "TestLibraryStringRefusals",
      "TestConsoleLowersToWriteLine",
      "TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat",
      "TestWhatZeroOneRefusesIsRefusedWithAFix",
      "TestAMutableLocationSeenWiderIsRefused",
      "TestAViewThatCantWriteIsNotRefused",
      "TestATupleSeenAsAnArrayIsNotYet",
      "TestAMethodReadAsAValueIsRefused"
    ],
    "evidence": "ADAMIC_MUTANT=M20 ADAMIC_BUILD_CACHE_DIR=/tmp/u036/cache/M20 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; lower_test.go:95: got /tmp/adamic-gate/TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhatan_arrow_r3582234619/001/main.a:3:2: stage 0 can't lower console.log with other than one string argument yet, want an error ending \"main.a:3:25: stage 0 can't lower a function returning never yet\"",
    "subsumption_mutants": null
  },
  {
    "test": "TestWhatZeroOneRefusesIsRefusedWithAFix",
    "package": "internal/lower",
    "file": "internal/lower/lower_test.go",
    "seconds": 0.483,
    "oracle": "Requires Refused and handwritten diagnostic substrings. Some expectations stop before the fix text, so the name promises more than every subcase checks.",
    "oracle_kind": "self",
    "kills": [
      "M11",
      "M13",
      "M15",
      "M20"
    ],
    "unique_kills": [],
    "last_proven_fail": "M20: lower_test.go:147: got /tmp/adamic-gate/TestWhatZeroOneRefusesIsRefusedWithAFixa_constructor_handing_th2532314978/001/main.a:12:1: stage 0 can't lower console.log with other than one string argument yet, want a refusal ending \"main.a:8:25: Adamic 0.1 refuses this escaping a constructor before every field is set (stored, passed, or a method called on it, which could read a field that holds undefined while its type says otherwise); assign every field first, then use this\"",
    "verdict": "overlapping",
    "subsumed_by": [
      "TestATupleSeenAsAnArrayIsNotYet",
      "TestAMethodReadAsAValueIsRefused"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": true,
    "matrix_rows": [
      "TestObjectRefusalsExplainSoundness",
      "TestObjectUnprovenShapesStayNotYet",
      "TestLibraryRegexOffsetRequiresNoCaptures",
      "TestLibraryStringRefusals",
      "TestConsoleLowersToWriteLine",
      "TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat",
      "TestWhatZeroOneRefusesIsRefusedWithAFix",
      "TestAMutableLocationSeenWiderIsRefused",
      "TestAViewThatCantWriteIsNotRefused",
      "TestATupleSeenAsAnArrayIsNotYet",
      "TestAMethodReadAsAValueIsRefused"
    ],
    "evidence": "ADAMIC_MUTANT=M20 ADAMIC_BUILD_CACHE_DIR=/tmp/u036/cache/M20 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; lower_test.go:147: got /tmp/adamic-gate/TestWhatZeroOneRefusesIsRefusedWithAFixa_constructor_handing_th2532314978/001/main.a:12:1: stage 0 can't lower console.log with other than one string argument yet, want a refusal ending \"main.a:8:25: Adamic 0.1 refuses this escaping a constructor before every field is set (stored, passed, or a method called on it, which could read a field that holds undefined while its type says otherwise); assign every field first, then use this\"",
    "subsumption_mutants": null
  },
  {
    "test": "TestAMutableLocationSeenWiderIsRefused",
    "package": "internal/lower",
    "file": "internal/lower/lower_test.go",
    "seconds": 0.557,
    "oracle": "Requires Refused and handwritten diagnostic substrings describing unsafe views. Does not execute the alleged unsound writes on Node.",
    "oracle_kind": "self",
    "kills": [
      "M10",
      "M11",
      "M13",
      "M15"
    ],
    "unique_kills": [],
    "last_proven_fail": "M15: lower_test.go:397: got /tmp/adamic-gate/TestAMutableLocationSeenWiderIsRefuseda_mutable_field_seen_as_a2234001085/001/main.a:18:19: Adamic 0.1 accepts a value of type Kennel seen as Pen, which can write Animal where Dog is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable), want a refusal starting \"main.a:18:19: Adamic 0.1 refuses a value of type Kennel seen as Pen, which can w [full output in log]",
    "verdict": "overlapping",
    "subsumed_by": [
      "TestATupleSeenAsAnArrayIsNotYet",
      "TestAMethodReadAsAValueIsRefused"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": true,
    "matrix_rows": [
      "TestObjectRefusalsExplainSoundness",
      "TestObjectUnprovenShapesStayNotYet",
      "TestLibraryRegexOffsetRequiresNoCaptures",
      "TestLibraryStringRefusals",
      "TestConsoleLowersToWriteLine",
      "TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat",
      "TestWhatZeroOneRefusesIsRefusedWithAFix",
      "TestAMutableLocationSeenWiderIsRefused",
      "TestAViewThatCantWriteIsNotRefused",
      "TestATupleSeenAsAnArrayIsNotYet",
      "TestAMethodReadAsAValueIsRefused"
    ],
    "evidence": "ADAMIC_MUTANT=M15 ADAMIC_BUILD_CACHE_DIR=/tmp/u036/cache/M15 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; lower_test.go:397: got /tmp/adamic-gate/TestAMutableLocationSeenWiderIsRefuseda_mutable_field_seen_as_a2234001085/001/main.a:18:19: Adamic 0.1 accepts a value of type Kennel seen as Pen, which can write Animal where Dog is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable), want a refusal starting \"main.a:18:19: Adamic 0.1 refuses a value of type Kennel seen as Pen, which can w [full output in log]",
    "subsumption_mutants": null
  },
  {
    "test": "TestAViewThatCantWriteIsNotRefused",
    "package": "internal/lower",
    "file": "internal/lower/lower_test.go",
    "seconds": 0.445,
    "oracle": "Rejects only errors.As(Refused). Accepts nil and all other errors, including NotYet. Its nil Lower entry probe passes all 26 subcases.",
    "oracle_kind": "self",
    "kills": [
      "M10",
      "M11",
      "M12",
      "M13"
    ],
    "unique_kills": [
      "M12"
    ],
    "last_proven_fail": "M13: lower_test.go:552: refused a view that can't write: /tmp/adamic-gate/TestAViewThatCantWriteIsNotRefusedan_array_seen_as_a_readonly_w1950604174/001/main.a:13:1: Adamic 0.1 refuses a method read as a value (log would lose its object, and this with it); call it in an arrow that keeps the object: (message) => console.log(message) (unbound-method)",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "vacuous_subcases": [
      "TestAViewThatCantWriteIsNotRefused/a_new_Map",
      "TestAViewThatCantWriteIsNotRefused/a_copy_made_by_map",
      "TestAViewThatCantWriteIsNotRefused/a_destructuring_declaration_of_a_readonly_field",
      "TestAViewThatCantWriteIsNotRefused/a_new_Map_passed",
      "TestAViewThatCantWriteIsNotRefused/a_union_target_that_can't_write",
      "TestAViewThatCantWriteIsNotRefused/a_conditional_of_fresh_copies",
      "TestAViewThatCantWriteIsNotRefused/a_method_returning_the_same_type",
      "TestAViewThatCantWriteIsNotRefused/a_class_seen_as_itself",
      "TestAViewThatCantWriteIsNotRefused/an_intersection_seen_as_itself",
      "TestAViewThatCantWriteIsNotRefused/an_intersection's_array_seen_as_a_readonly_array_with_the_same_tag",
      "TestAViewThatCantWriteIsNotRefused/a_readonly_field_copied_into_a_writable_one",
      "TestAViewThatCantWriteIsNotRefused/a_copy_made_by_slice",
      "TestAViewThatCantWriteIsNotRefused/a_writable_field_seen_as_a_readonly_one",
      "TestAViewThatCantWriteIsNotRefused/a_readonly_field_seen_as_its_own_type",
      "TestAViewThatCantWriteIsNotRefused/a_shorthand_property_that_can't_write",
      "TestAViewThatCantWriteIsNotRefused/a_readonly_field_seen_as_a_readonly_field",
      "TestAViewThatCantWriteIsNotRefused/a_readonly_constraint:_a_Narrow_into_a_Pack_slot",
      "TestAViewThatCantWriteIsNotRefused/a_type_parameter_narrowed_from_undefined,_as_itself",
      "TestAViewThatCantWriteIsNotRefused/a_destructuring_with_no_type",
      "TestAViewThatCantWriteIsNotRefused/a_type_parameter_as_itself",
      "TestAViewThatCantWriteIsNotRefused/a_Map_made_from_pairs",
      "TestAViewThatCantWriteIsNotRefused/a_conditional_of_fresh_arrays",
      "TestAViewThatCantWriteIsNotRefused/an_array_seen_as_a_readonly_wider_array",
      "TestAViewThatCantWriteIsNotRefused/a_copy_made_by_filter",
      "TestAViewThatCantWriteIsNotRefused/a_readonly_field_seen_wider",
      "TestAViewThatCantWriteIsNotRefused/a_destructuring_assignment's_elements_going_into_wider_names"
    ],
    "bounded": true,
    "matrix_rows": [
      "TestObjectRefusalsExplainSoundness",
      "TestObjectUnprovenShapesStayNotYet",
      "TestLibraryRegexOffsetRequiresNoCaptures",
      "TestLibraryStringRefusals",
      "TestConsoleLowersToWriteLine",
      "TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat",
      "TestWhatZeroOneRefusesIsRefusedWithAFix",
      "TestAMutableLocationSeenWiderIsRefused",
      "TestAViewThatCantWriteIsNotRefused",
      "TestATupleSeenAsAnArrayIsNotYet",
      "TestAMethodReadAsAValueIsRefused"
    ],
    "evidence": "ADAMIC_MUTANT=M13 ADAMIC_BUILD_CACHE_DIR=/tmp/u036/cache/M13 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; lower_test.go:552: refused a view that can't write: /tmp/adamic-gate/TestAViewThatCantWriteIsNotRefusedan_array_seen_as_a_readonly_w1950604174/001/main.a:13:1: Adamic 0.1 refuses a method read as a value (log would lose its object, and this with it); call it in an arrow that keeps the object: (message) => console.log(message) (unbound-method)",
    "subsumption_mutants": null
  },
  {
    "test": "TestATupleSeenAsAnArrayIsNotYet",
    "package": "internal/lower",
    "file": "internal/lower/lower_test.go",
    "seconds": 0.301,
    "oracle": "Requires NotYet and handwritten diagnostic suffixes naming representation mismatch and workaround.",
    "oracle_kind": "self",
    "kills": [
      "M10",
      "M11",
      "M13",
      "M16",
      "M20"
    ],
    "unique_kills": [],
    "last_proven_fail": "M20: lower_test.go:589: got /tmp/adamic-gate/TestATupleSeenAsAnArrayIsNotYetan_argument886451439/001/main.a:5:1: stage 0 can't lower console.log with other than one string argument yet, want a not-yet ending \"main.a:5:22: stage 0 can't lower a [number, number] seen as a readonly number[] (a tuple is held as an object, not an array, so far; write it as an array where it's made, or copy it into one: [pair[0], pair[1]]) yet\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 0.478,
    "vacuous": false,
    "vacuous_subcases": [],
    "bounded": true,
    "matrix_rows": [
      "TestObjectRefusalsExplainSoundness",
      "TestObjectUnprovenShapesStayNotYet",
      "TestLibraryRegexOffsetRequiresNoCaptures",
      "TestLibraryStringRefusals",
      "TestConsoleLowersToWriteLine",
      "TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat",
      "TestWhatZeroOneRefusesIsRefusedWithAFix",
      "TestAMutableLocationSeenWiderIsRefused",
      "TestAViewThatCantWriteIsNotRefused",
      "TestATupleSeenAsAnArrayIsNotYet",
      "TestAMethodReadAsAValueIsRefused"
    ],
    "evidence": "ADAMIC_MUTANT=M20 ADAMIC_BUILD_CACHE_DIR=/tmp/u036/cache/M20 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; lower_test.go:589: got /tmp/adamic-gate/TestATupleSeenAsAnArrayIsNotYetan_argument886451439/001/main.a:5:1: stage 0 can't lower console.log with other than one string argument yet, want a not-yet ending \"main.a:5:22: stage 0 can't lower a [number, number] seen as a readonly number[] (a tuple is held as an object, not an array, so far; write it as an array where it's made, or copy it into one: [pair[0], pair[1]]) yet\"",
    "subsumption_mutants": 5
  },
  {
    "test": "TestAMethodReadAsAValueIsRefused",
    "package": "internal/lower",
    "file": "internal/lower/lower_test.go",
    "seconds": 0.197,
    "oracle": "Negative cases require Refused and diagnostic substrings; four positive neighbors reject only Refused and accept an empty Lower answer.",
    "oracle_kind": "self",
    "kills": [
      "M13",
      "M15"
    ],
    "unique_kills": [],
    "last_proven_fail": "M15: lower_test.go:629: got /tmp/adamic-gate/TestAMethodReadAsAValueIsRefusedpassed_as_a_callback3306394704/001/main.a:11:25: Adamic 0.1 accepts a method read as a value (admit would lose its object, and this with it); call it in an arrow that keeps the object: (pet) => shelter.admit(pet) (unbound-method), want a refusal containing \"main.a:11:25: Adamic 0.1 refuses a method read as a value (admit would lose its object\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestLibraryStringRefusals"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": 0.21,
    "vacuous": false,
    "vacuous_subcases": [
      "TestAMethodReadAsAValueIsRefused/not_called_through_parentheses",
      "TestAMethodReadAsAValueIsRefused/not_called_in_an_arrow",
      "TestAMethodReadAsAValueIsRefused/not_a_field_holding_a_function",
      "TestAMethodReadAsAValueIsRefused/not_called_on_its_object"
    ],
    "bounded": true,
    "matrix_rows": [
      "TestObjectRefusalsExplainSoundness",
      "TestObjectUnprovenShapesStayNotYet",
      "TestLibraryRegexOffsetRequiresNoCaptures",
      "TestLibraryStringRefusals",
      "TestConsoleLowersToWriteLine",
      "TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat",
      "TestWhatZeroOneRefusesIsRefusedWithAFix",
      "TestAMutableLocationSeenWiderIsRefused",
      "TestAViewThatCantWriteIsNotRefused",
      "TestATupleSeenAsAnArrayIsNotYet",
      "TestAMethodReadAsAValueIsRefused"
    ],
    "evidence": "ADAMIC_MUTANT=M15 ADAMIC_BUILD_CACHE_DIR=/tmp/u036/cache/M15 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .; lower_test.go:629: got /tmp/adamic-gate/TestAMethodReadAsAValueIsRefusedpassed_as_a_callback3306394704/001/main.a:11:25: Adamic 0.1 accepts a method read as a value (admit would lose its object, and this with it); call it in an arrow that keeps the object: (pet) => shelter.admit(pet) (unbound-method), want a refusal containing \"main.a:11:25: Adamic 0.1 refuses a method read as a value (admit would lose its object\"",
    "subsumption_mutants": 2
  }
]
```

| ID | origin/main location | Change | Failed assigned rows | Full-package wall seconds |
|---|---|---|---|---:|
| M01 | internal/lower/library_object.go:17 | `Fix: reason` to `Fix: ""` | TestObjectRefusalsExplainSoundness | 34.562 |
| M02 | internal/lower/library_object.go:29 | `count := 1` to `count := 2` | TestObjectRefusalsExplainSoundness | 29.489 |
| M03 | internal/lower/library_object.go:197 | `if depth > 16 {` to `if depth > 0 {` |  | 29.360 |
| M04 | internal/lower/library_object.go:92 | `!l.hasProperty(written[0], key.Text())` to `l.hasProperty(written[0], key.Text())` | TestObjectRefusalsExplainSoundness | 31.252 |
| M05 | internal/lower/library_object.go:27 | `Object.groupBy's partial record` to `Object.collectBy's partial record` |  | 28.709 |
| M06 | internal/lower/library_object.go:230 | `declaration.Parent.Flags&ast.NodeFlagsConst == 0` to `declaration.Parent.Flags&ast.NodeFlagsConst != 0` |  | 28.176 |
| M07 | internal/lower/library_regex_callback_shape.go:58 | `n.Kind != regex.Capturing` to `n.Kind == regex.Capturing` | TestLibraryRegexOffsetRequiresNoCaptures | 36.313 |
| M08 | internal/lower/library_regex_callback_shape.go:13 | `if depth > 32 {` to `if depth > 0 {` |  | 29.071 |
| M09 | internal/lower/library_regex_callback_shape.go:32 | `declaration.Parent.Flags&ast.NodeFlagsConst == 0` to `declaration.Parent.Flags&ast.NodeFlagsConst != 0` | TestLibraryRegexOffsetRequiresNoCaptures | 30.114 |
| M10 | internal/lower/invariance.go:211 | `mutable := !l.isLibraryType(to, "ReadonlyArray", "ReadonlyMap", "ReadonlySet")` to `mutable := l.isLibraryType(to, "ReadonlyArray", "ReadonlyMap", "ReadonlySet")` | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat, TestAMutableLocationSeenWiderIsRefused, TestAViewThatCantWriteIsNotRefused, TestATupleSeenAsAnArrayIsNotYet | 30.580 |
| M11 | internal/lower/invariance.go:181 | `if !fresh && !l.checker.IsReadonlySymbol(viewed) && l.checker.IsReadonlySymbol(inside) {` to `if !fresh && l.checker.IsReadonlySymbol(viewed) && l.checker.IsReadonlySymbol(inside) {` | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat, TestWhatZeroOneRefusesIsRefusedWithAFix, TestAMutableLocationSeenWiderIsRefused, TestAViewThatCantWriteIsNotRefused, TestATupleSeenAsAnArrayIsNotYet | 27.833 |
| M12 | internal/lower/invariance.go:593 | `return l.checker.IsArrayType(l.checker.GetTypeAtLocation(access.Expression)) || l.isLibraryType(l.checker.GetTypeAtLocation(access.Expression), "ReadonlyArray")` to `return false` | TestAViewThatCantWriteIsNotRefused | 31.880 |
| M13 | internal/lower/refusals.go:183 | `if node.Kind == ast.KindPropertyAccessExpression && !called(node) && !l.libraryNumberBoundMethod(node)` to `if node.Kind == ast.KindPropertyAccessExpression && called(node) && !l.libraryNumberBoundMethod(node)` | TestObjectRefusalsExplainSoundness, TestObjectUnprovenShapesStayNotYet, TestLibraryStringRefusals, TestConsoleLowersToWriteLine, TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat, TestWhatZeroOneRefusesIsRefusedWithAFix, TestAMutableLocationSeenWiderIsRefused, TestAViewThatCantWriteIsNotRefused, TestATupleSeenAsAnArrayIsNotYet, TestAMethodReadAsAValueIsRefused | 36.947 |
| M14 | internal/lower/prelude.go:72 | `return ir.Stderr, nil` to `return ir.Stdout, nil` | TestConsoleLowersToWriteLine | 42.904 |
| M15 | internal/lower/diagnostics.go:29 | `%s: Adamic 0.1 refuses %s; %s` to `%s: Adamic 0.1 accepts %s; %s` | TestLibraryStringRefusals, TestWhatZeroOneRefusesIsRefusedWithAFix, TestAMutableLocationSeenWiderIsRefused, TestAMethodReadAsAValueIsRefused | 45.998 |
| M16 | internal/lower/diagnostics.go:17 | `%s: stage 0 can't lower %s yet` to `%s: stage 0 can't lower %s now` | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat, TestATupleSeenAsAnArrayIsNotYet | 43.262 |
| M17 | internal/lower/object.go:2078 | `if value == nil || target == nil || value == target || depth > 4 {` to `if value == nil || target == nil || value == target || depth > 0 {` |  | 42.506 |
| M18 | internal/lower/statements.go:68 | `return nil, l.notYet(node, "a class inside a function")` to `return nil, l.notYet(node, "a nested class")` | TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat | 43.819 |
| M19 | internal/lower/library_object.go:241 | `call.Method == "freeze"` to `call.Method != "freeze"` | TestObjectUnprovenShapesStayNotYet | 33.961 |
| M20 | internal/lower/prelude.go:17 | `if len(arguments) != 1 {` to `if len(arguments) == 1 {` | TestLibraryStringRefusals, TestConsoleLowersToWriteLine, TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat, TestWhatZeroOneRefusesIsRefusedWithAFix, TestATupleSeenAsAnArrayIsNotYet | 28.247 |
| P01 | internal/lower/lower.go:20 | `func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {` to `return nil, nil` | TestObjectRefusalsExplainSoundness, TestObjectUnprovenShapesStayNotYet, TestLibraryStringRefusals, TestConsoleLowersToWriteLine, TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat, TestWhatZeroOneRefusesIsRefusedWithAFix, TestAMutableLocationSeenWiderIsRefused, TestATupleSeenAsAnArrayIsNotYet, TestAMethodReadAsAValueIsRefused | 1.730 |
| P02 | internal/lower/library_regex_callback_shape.go:12 | `func (l *lowering) regexCallbackNoCaptures(node *ast.Node, depth int) bool {` to `return false` |  | 28.345 |

Survivors:
- M03 is a slice survivor, caught by TestNodeFSFileOptionsBorrow and TestNodeFSFileScratchOptionsBorrow outside the slice. The independent const-freeze caller changes from successful lowering to NotYet about an unproven shape.
- M06 is a slice survivor, caught by those same two outside rows. The const-freeze caller also changes success to NotYet.
- M05 survives the entire default package. The same Object.groupBy input produces a diagnostic naming Object.collectBy instead. TestObjectUnprovenShapesStayNotYet only checks the NotYet class and accepts both.
- M08 survives the entire default package. The output-only internal observer prints const regex capture-free proof true before, false after. The corresponding standalone Lower caller still succeeds in both runs. This proves changed internal query behavior; it does not demonstrate a changed native product.
- M17 survives the entire default package. The output-only internal observer prints nested tuple-to-array query true before, false after. Both attempted standalone Lower callers still report NotYet through other checks. This proves changed query behavior, with observable lowering consequences not established.

Witness commands: go build -o /tmp/u036-witness review/test-audit/internal-lower-library_object/survivor-witness.go, then ADAMIC_MUTANT=<id> /tmp/u036-witness while the switch is installed. Raw outputs are witness-*.log. witness-M02.log also records the wrong argument-count NotYet reason, documenting the class-only oracle weakness. For unexported queries, the output-only scratch TestU036QueryWitness in query-witness.go.txt was temporarily added only after the main matrix finished. Command: ADAMIC_MUTANT=<id> ADAMIC_BUILD_CACHE_DIR=/tmp/u036/cache/<id>-query timeout 120 go test -v -count=1 -timeout 90s ./internal/lower/ -run '^TestU036QueryWitness$'. It prints observations without comparing against an expected answer, has no verdict, and is excluded from the matrix. query-clean.log, query-M08.log, and query-M17.log show the changed outputs.

Procedure, unclear points, and limits:
- The brief names files at 8de93800f4 but requires fetching current origin/main. Current origin/main was e2492670b06a4dce1deafe837158ce0366cf2bc4. All 11 names remained in the four named files; none moved or vanished. Every diff and location uses this actual starting commit.
- All assigned tests are independent rows. They share preparation through lowerSource, but check different assertions. There are no assigned families, setup checks, helpers, or witnesses.
- The complete pre-mutation inventory contains 447 coverage-observed functions in internal/lower. It was saved before selecting and testing production mutations. That does not mean all 447 functions were mutated or exhaustively analyzed. The 20 mutations cover nine files and several distinct production functions.
- Whole-package discovery and baseline succeeded. Every mutant was run against the whole package, because the baseline fit 90 seconds. The reported verdicts and unique_kills are nevertheless bounded to the 11 assigned rows, as requested for a slice. Other failure names are retained in mutants.json and raw logs, but out-of-scope families were not adjudicated. No repo-wide uniqueness claim is made.
- All 20 final standalone production diffs and both probe diffs passed go vet ./internal/lower/. No compile warning or build failure is counted as a kill. A draft M20 left stream unused and failed vet; it was replaced with an argument-count condition flip before any production matrix observations. The invalid draft is saved as .diff.txt rather than a replayable .diff.
- The initial generic switch applied M19 to both matching freeze callbacks, but M19.diff changes only the first. The initial run is preserved as M19-two-occurrences.log and excluded from verdicts. Exact standalone M19 was rerun over the whole package with /tmp/u036/cache/M19-corrected, confirming the assigned catch. switch.diff was aligned to the single-occurrence mutation; the initial switch is preserved as switch-initial.diff.txt.
- Every compiler matrix run used ADAMIC_BUILD_CACHE_DIR=/tmp/u036/cache/<mutant>, with a distinct corrected-replay cache for M19. Native build-only times were not separately instrumented; run-meta.json records each full command's wall duration, including compilation, native construction and running. The slice's rows themselves perform no native comparison.
- P01 returns nil,nil at Lower entry; P02 returns false at regexCallbackNoCaptures entry. These are probes, not production kills. P01 aborts the whole binary, so every assigned row was rerun alone. Only those observations decide P01 results. P02 passes the whole default package. There are no unknown assigned matrix cells and no test-budget timeouts.
- TestAViewThatCantWriteIsNotRefused accepts every non-Refused error and nil output. All 26 subcases pass P01. TestLibraryRegexOffsetRequiresNoCaptures contains only rejecting proof inputs and passes P02. Both are vacuous findings even though eligible mutants establish bounded sacred verdicts.
- TestAMethodReadAsAValueIsRefused is not vacuous overall: its negative cases fail P01. Its four positive neighbors pass P01 and are listed as vacuous_subcases. Passing subcases for other probes are retained in rows.json and raw logs.
- TestObjectUnprovenShapesStayNotYet checks only the error class. It cannot distinguish the incorrect diagnostic from M05, and several wrong-argument-count NotYet answers from M02 still pass its subcases. This is an oracle-strength finding, not an external-authority claim.
- TestWhatZeroOneRefusesIsRefusedWithAFix has expected substrings which stop before the fix in some cases. Its name does not prove every fix text is checked. All 11 oracle_kind values are self. Calls to the in-process checker prepare inputs; these rows do not compare against Node, Go cohere, or an outside specification.
- Subsumption rests on 3 catches for TestLibraryStringRefusals, 5 for TestATupleSeenAsAnArrayIsNotYet, and 2 for TestAMethodReadAsAValueIsRefused. The named subsumers have medians 0.483, 0.478, and 0.210 seconds respectively. These bounded hints do not authorize deletion.
- Default whole-package baseline skips outside the assigned slice: TestOriginalCycleLedger (requires a separately pinned pristine TypeScript corpus), TestOptionalWideningCensus (requires a user-selected project/config and output destination), and one TestMixedUnionContractGraph subcase. No assigned row skipped. Those unrelated opt-in inventories were not enabled for this slice; their absence limits default-package survivor claims.
- The package's full baseline is not red. Some native runtime tests outside the slice were rebuilt by full matrix runs; no other package's tests or repository-wide gate was run. Go vet may compile dependencies without running their tests.
- Production source was restored exactly, the scratch observation test was removed, git diff --check passed, and the assigned clean slice passed again in 2.194 seconds. Logs are force-added because repository ignore rules ignore *.log.

Timing and coverage:
- Warm env.sh worked with Go 1.27.1; setup.sh was skipped. npm ci completed in 0.519 seconds, nproc 5.
- Whole clean package: 34.628 binary seconds. Clean slice with coverage: 2.240 seconds, 28.7 percent statement coverage.
- Thirty-three isolated timing commands: 63.398 wall seconds. Reported row medians come from each test binary's own ok line, not shell overhead.
- Twenty final standalone mutant vet commands: 9.274 wall seconds. Successful scratch switch gofmt/vet: 0.469 seconds. Both entry-probe standalone diffs were additionally vetted.
- Mutation/probe commands, isolated P01 recovery, and both M19 runs: 780.074 wall seconds. Exact standalone M19 replay: 33.961 seconds. Per-mutant full command times appear in the table and run-meta.json. Native build-only time is not separated from test running.
- Total session effort exceeded the approximately 20-minute target because full-package matrices and replay verification took longer than the bounded slice alone. No individual test binary exceeded 90 seconds.
- Not covered: repo-wide replay, adjudication of out-of-scope families/rows, external expected-answer checks, and proof that M08/M17 alter emitted/native behavior. Their internal query outputs did change; they are not labeled equivalent candidates.
