u057 audited all 13 requested rows at 68db8ddd145281a62655452496bdc32ef848bdf3; none moved or vanished.
Whole-package baseline timed out at 90 seconds; clean bounded baseline and final rerun passed.
Four vetted production mutants: two caught and two survived the bounded matrix; no unique production kills.
Six witnesses proved; two witness rows and one production row are untrue under this experiment.
Five production rows failed their empty-entry probes; evidence is on the requested audit branch.

```json
[
  {
    "test": "TestCheckedViewFlagDowncast",
    "package": "internal/oracle",
    "file": "internal/oracle/checked_views_flag_downcast_test.go:12",
    "seconds": 0.414,
    "oracle": "Node source execution plus handwritten boundary/panic text and sanitizer/leak checks; full output comparison",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: checked_views_flag_downcast_test.go:70: native release field-check mutant survived",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCheckedViewFlagDowncast",
      "TestCheckedViewInterfaces",
      "TestCheckedViewObjects",
      "TestCheckedViewOptionalReadBoundary",
      "TestCheckedViewV2ReadsAfterWrites",
      "TestCheckedViewV2ArrayArmBoundary",
      "TestCheckedViewV2WrongFamilyMutant",
      "TestCheckedViewV2RepresentationMutants",
      "TestCheckedViewV2MembershipMutant",
      "TestCheckedViewV2TupleIdentityMutant",
      "TestCheckedViewV2CallableProducerMutant",
      "TestCheckedViewV2MovedStage3Results",
      "TestCheckedViewObjectPrimitiveSource"
    ],
    "evidence": "ADAMIC_MUTANT=W1 ADAMIC_BUILD_CACHE_DIR=/tmp/u057/cache/W1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestCheckedViewFlagDowncast|TestCheckedViewInterfaces|TestCheckedViewObjects|TestCheckedViewV2WrongFamilyMutant|TestCheckedViewV2RepresentationMutants|TestCheckedViewV2MembershipMutant|TestCheckedViewV2TupleIdentityMutant|TestCheckedViewV2CallableProducerMutant)$' > W1.log 2>&1; checked_views_flag_downcast_test.go:70: native release field-check mutant survived",
    "witness": true
  },
  {
    "test": "TestCheckedViewInterfaces",
    "package": "internal/oracle",
    "file": "internal/oracle/checked_views_interfaces_test.go:5",
    "seconds": 0.465,
    "oracle": "Node source execution plus handwritten boundary/panic text and sanitizer/leak checks; full output comparison",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: --- FAIL: TestCheckedViewInterfaces/interfaces-wrong-inherited (0.70s)",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCheckedViewFlagDowncast",
      "TestCheckedViewInterfaces",
      "TestCheckedViewObjects",
      "TestCheckedViewOptionalReadBoundary",
      "TestCheckedViewV2ReadsAfterWrites",
      "TestCheckedViewV2ArrayArmBoundary",
      "TestCheckedViewV2WrongFamilyMutant",
      "TestCheckedViewV2RepresentationMutants",
      "TestCheckedViewV2MembershipMutant",
      "TestCheckedViewV2TupleIdentityMutant",
      "TestCheckedViewV2CallableProducerMutant",
      "TestCheckedViewV2MovedStage3Results",
      "TestCheckedViewObjectPrimitiveSource"
    ],
    "evidence": "ADAMIC_MUTANT=W1 ADAMIC_BUILD_CACHE_DIR=/tmp/u057/cache/W1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestCheckedViewFlagDowncast|TestCheckedViewInterfaces|TestCheckedViewObjects|TestCheckedViewV2WrongFamilyMutant|TestCheckedViewV2RepresentationMutants|TestCheckedViewV2MembershipMutant|TestCheckedViewV2TupleIdentityMutant|TestCheckedViewV2CallableProducerMutant)$' > W1.log 2>&1; --- FAIL: TestCheckedViewInterfaces/interfaces-wrong-inherited (0.70s)",
    "witness": true
  },
  {
    "test": "TestCheckedViewObjects",
    "package": "internal/oracle",
    "file": "internal/oracle/checked_views_objects_test.go:9",
    "seconds": 0.782,
    "oracle": "Node source execution plus handwritten boundary/panic text and sanitizer/leak checks; full output comparison",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: --- FAIL: TestCheckedViewObjects/objects-wrong-nested (0.71s)",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCheckedViewFlagDowncast",
      "TestCheckedViewInterfaces",
      "TestCheckedViewObjects",
      "TestCheckedViewOptionalReadBoundary",
      "TestCheckedViewV2ReadsAfterWrites",
      "TestCheckedViewV2ArrayArmBoundary",
      "TestCheckedViewV2WrongFamilyMutant",
      "TestCheckedViewV2RepresentationMutants",
      "TestCheckedViewV2MembershipMutant",
      "TestCheckedViewV2TupleIdentityMutant",
      "TestCheckedViewV2CallableProducerMutant",
      "TestCheckedViewV2MovedStage3Results",
      "TestCheckedViewObjectPrimitiveSource"
    ],
    "evidence": "ADAMIC_MUTANT=W1 ADAMIC_BUILD_CACHE_DIR=/tmp/u057/cache/W1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestCheckedViewFlagDowncast|TestCheckedViewInterfaces|TestCheckedViewObjects|TestCheckedViewV2WrongFamilyMutant|TestCheckedViewV2RepresentationMutants|TestCheckedViewV2MembershipMutant|TestCheckedViewV2TupleIdentityMutant|TestCheckedViewV2CallableProducerMutant)$' > W1.log 2>&1; --- FAIL: TestCheckedViewObjects/objects-wrong-nested (0.71s)",
    "witness": true
  },
  {
    "test": "TestCheckedViewOptionalReadBoundary",
    "package": "internal/oracle",
    "file": "internal/oracle/checked_views_optional_read_test.go:16",
    "seconds": 0.352,
    "oracle": "Node source execution plus handwritten boundary/panic text and sanitizer/leak checks; full output comparison",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M2"
    ],
    "unique_kills": [],
    "last_proven_fail": "M2: checked_views_optional_read_test.go:65: boundary diagnostic: got \"/workspace/adamic/stage3/interface-downcasts/lane1/optional-read.a:9:13: stage 0 can't lower a checked field alias requiring an optional, accessor, or representation adaptation yet\", want \"/workspace/adamic/stage3/interface-downcasts/lane1/optional-read.a:9:13: stage 0 can't lower a checked field alias requiring an optional, accessor, or representation conversion yet\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestCheckedViewObjectPrimitiveSource"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "PLower",
      "PC",
      "PJS"
    ],
    "subsumer_seconds": 1.604,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestCheckedViewFlagDowncast",
      "TestCheckedViewInterfaces",
      "TestCheckedViewObjects",
      "TestCheckedViewOptionalReadBoundary",
      "TestCheckedViewV2ReadsAfterWrites",
      "TestCheckedViewV2ArrayArmBoundary",
      "TestCheckedViewV2WrongFamilyMutant",
      "TestCheckedViewV2RepresentationMutants",
      "TestCheckedViewV2MembershipMutant",
      "TestCheckedViewV2TupleIdentityMutant",
      "TestCheckedViewV2CallableProducerMutant",
      "TestCheckedViewV2MovedStage3Results",
      "TestCheckedViewObjectPrimitiveSource"
    ],
    "evidence": "ADAMIC_MUTANT=M2 ADAMIC_BUILD_CACHE_DIR=/tmp/u057/cache/M2 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestCheckedViewFlagDowncast|TestCheckedViewInterfaces|TestCheckedViewObjects|TestCheckedViewOptionalReadBoundary|TestCheckedViewV2ReadsAfterWrites|TestCheckedViewV2ArrayArmBoundary|TestCheckedViewV2WrongFamilyMutant|TestCheckedViewV2RepresentationMutants|TestCheckedViewV2MembershipMutant|TestCheckedViewV2TupleIdentityMutant|TestCheckedViewV2CallableProducerMutant|TestCheckedViewV2MovedStage3Results|TestCheckedViewObjectPrimitiveSource)$' > M2.log 2>&1; checked_views_optional_read_test.go:65: boundary diagnostic: got \"/workspace/adamic/stage3/interface-downcasts/lane1/optional-read.a:9:13: stage 0 can't lower a checked field alias requiring an optional, accessor, or representation adaptation yet\", want \"/workspace/adamic/stage3/interface-downcasts/lane1/optional-read.a:9:13: stage 0 can't lower a checked field alias requiring an optional, accessor, or representation conversion yet\"",
    "witness": false,
    "vacuous_subcases": "PC/PJS preserve lowering-refusal cases which do not reach those entries; no positive executed case survived its Lower probe."
  },
  {
    "test": "TestCheckedViewV2ReadsAfterWrites",
    "package": "internal/oracle",
    "file": "internal/oracle/checked_views_v2_migration_test.go:15",
    "seconds": 0.852,
    "oracle": "Node source execution compared with native and JavaScript outputs, plus sanitizer/leak checks",
    "oracle_kind": "external-run",
    "kills": [
      "M3"
    ],
    "unique_kills": [],
    "last_proven_fail": "M3: checked_views_v2_migration_test.go:25: native: exit codes differ; oracle.run{stdout:[]uint8{}, stderr:[]uint8{0x61, 0x64, 0x61, 0x6d, 0x69, 0x63, 0x3a, 0x20, 0x70, 0x61, 0x6e, 0x69, 0x63, 0x3a, 0x20, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x20, 0x72, 0x65, 0x61, 0x64, 0x20, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x3a, 0x20, 0x76, 0x69, 0x65, 0x77, 0x65, 0x64, 0x2e, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x5b, 0x30, 0x5d, 0x20, 0x69, 0x73, 0x20, 0x6e, 0x6f, 0x74, 0x20, 0x61, 0x20, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x3b, 0x20, 0x65, 0x78, 0x70, 0x65, 0x63, 0x74, 0x65, 0x64, 0x20, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x2c, 0x20, 0x66, 0x6f, 0x75, 0x6e, 0x64, 0x20, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0xa}, exitCode:70}",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestCheckedViewV2MovedStage3Results"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "PLower",
      "PC",
      "PJS"
    ],
    "subsumer_seconds": 0.693,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestCheckedViewFlagDowncast",
      "TestCheckedViewInterfaces",
      "TestCheckedViewObjects",
      "TestCheckedViewOptionalReadBoundary",
      "TestCheckedViewV2ReadsAfterWrites",
      "TestCheckedViewV2ArrayArmBoundary",
      "TestCheckedViewV2WrongFamilyMutant",
      "TestCheckedViewV2RepresentationMutants",
      "TestCheckedViewV2MembershipMutant",
      "TestCheckedViewV2TupleIdentityMutant",
      "TestCheckedViewV2CallableProducerMutant",
      "TestCheckedViewV2MovedStage3Results",
      "TestCheckedViewObjectPrimitiveSource"
    ],
    "evidence": "ADAMIC_MUTANT=M3 ADAMIC_BUILD_CACHE_DIR=/tmp/u057/cache/M3 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestCheckedViewFlagDowncast|TestCheckedViewInterfaces|TestCheckedViewObjects|TestCheckedViewOptionalReadBoundary|TestCheckedViewV2ReadsAfterWrites|TestCheckedViewV2ArrayArmBoundary|TestCheckedViewV2WrongFamilyMutant|TestCheckedViewV2RepresentationMutants|TestCheckedViewV2MembershipMutant|TestCheckedViewV2TupleIdentityMutant|TestCheckedViewV2CallableProducerMutant|TestCheckedViewV2MovedStage3Results|TestCheckedViewObjectPrimitiveSource)$' > M3.log 2>&1; checked_views_v2_migration_test.go:25: native: exit codes differ; oracle.run{stdout:[]uint8{}, stderr:[]uint8{0x61, 0x64, 0x61, 0x6d, 0x69, 0x63, 0x3a, 0x20, 0x70, 0x61, 0x6e, 0x69, 0x63, 0x3a, 0x20, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x20, 0x72, 0x65, 0x61, 0x64, 0x20, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x3a, 0x20, 0x76, 0x69, 0x65, 0x77, 0x65, 0x64, 0x2e, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x5b, 0x30, 0x5d, 0x20, 0x69, 0x73, 0x20, 0x6e, 0x6f, 0x74, 0x20, 0x61, 0x20, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x3b, 0x20, 0x65, 0x78, 0x70, 0x65, 0x63, 0x74, 0x65, 0x64, 0x20, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x2c, 0x20, 0x66, 0x6f, 0x75, 0x6e, 0x64, 0x20, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0xa}, exitCode:70}",
    "witness": false
  },
  {
    "test": "TestCheckedViewV2ArrayArmBoundary",
    "package": "internal/oracle",
    "file": "internal/oracle/checked_views_v2_migration_test.go:35",
    "seconds": 0.12,
    "oracle": "Node success exit only; handwritten views-v3 refusal substring. The source check can pass for a wrong successful output.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "PLower"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestCheckedViewFlagDowncast",
      "TestCheckedViewInterfaces",
      "TestCheckedViewObjects",
      "TestCheckedViewOptionalReadBoundary",
      "TestCheckedViewV2ReadsAfterWrites",
      "TestCheckedViewV2ArrayArmBoundary",
      "TestCheckedViewV2WrongFamilyMutant",
      "TestCheckedViewV2RepresentationMutants",
      "TestCheckedViewV2MembershipMutant",
      "TestCheckedViewV2TupleIdentityMutant",
      "TestCheckedViewV2CallableProducerMutant",
      "TestCheckedViewV2MovedStage3Results",
      "TestCheckedViewObjectPrimitiveSource"
    ],
    "evidence": "ADAMIC_MUTANT=None ADAMIC_BUILD_CACHE_DIR=/tmp/u057/cache/None timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestCheckedViewFlagDowncast|TestCheckedViewInterfaces|TestCheckedViewObjects|TestCheckedViewOptionalReadBoundary|TestCheckedViewV2ReadsAfterWrites|TestCheckedViewV2ArrayArmBoundary|TestCheckedViewV2WrongFamilyMutant|TestCheckedViewV2RepresentationMutants|TestCheckedViewV2MembershipMutant|TestCheckedViewV2TupleIdentityMutant|TestCheckedViewV2CallableProducerMutant|TestCheckedViewV2MovedStage3Results|TestCheckedViewObjectPrimitiveSource)$' > M1.log 2>&1; PASS observed",
    "witness": false
  },
  {
    "test": "TestCheckedViewV2WrongFamilyMutant",
    "package": "internal/oracle",
    "file": "internal/oracle/checked_views_v2_migration_test.go:51",
    "seconds": 0.489,
    "oracle": "Node source execution plus handwritten boundary/panic text and sanitizer/leak checks; full output comparison",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: checked_views_v2_migration_test.go:82: wrong adapter survived javascript: oracle.run{stdout:[]uint8{}, stderr:[]uint8{0x61, 0x64, 0x61, 0x6d, 0x69, 0x63, 0x3a, 0x20, 0x70, 0x61, 0x6e, 0x69, 0x63, 0x3a, 0x20, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x20, 0x72, 0x65, 0x61, 0x64, 0x20, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x3a, 0x20, 0x76, 0x69, 0x65, 0x77, 0x2e, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x20, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x65, 0x73, 0x20, 0x6e, 0x6f, 0x20, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x20, 0x6f, 0x66, 0x20, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74, 0x3b, 0x20, 0x65, 0x78, 0x70, 0x65, 0x63, 0x74, 0x65, 0x64, 0x20, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74, 0x2c, 0x20, 0x66, 0x6f, 0x75, 0x6e, 0x64, 0x20, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0xa}, exitCode:70}",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCheckedViewFlagDowncast",
      "TestCheckedViewInterfaces",
      "TestCheckedViewObjects",
      "TestCheckedViewOptionalReadBoundary",
      "TestCheckedViewV2ReadsAfterWrites",
      "TestCheckedViewV2ArrayArmBoundary",
      "TestCheckedViewV2WrongFamilyMutant",
      "TestCheckedViewV2RepresentationMutants",
      "TestCheckedViewV2MembershipMutant",
      "TestCheckedViewV2TupleIdentityMutant",
      "TestCheckedViewV2CallableProducerMutant",
      "TestCheckedViewV2MovedStage3Results",
      "TestCheckedViewObjectPrimitiveSource"
    ],
    "evidence": "ADAMIC_MUTANT=W1 ADAMIC_BUILD_CACHE_DIR=/tmp/u057/cache/W1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestCheckedViewFlagDowncast|TestCheckedViewInterfaces|TestCheckedViewObjects|TestCheckedViewV2WrongFamilyMutant|TestCheckedViewV2RepresentationMutants|TestCheckedViewV2MembershipMutant|TestCheckedViewV2TupleIdentityMutant|TestCheckedViewV2CallableProducerMutant)$' > W1.log 2>&1; checked_views_v2_migration_test.go:82: wrong adapter survived javascript: oracle.run{stdout:[]uint8{}, stderr:[]uint8{0x61, 0x64, 0x61, 0x6d, 0x69, 0x63, 0x3a, 0x20, 0x70, 0x61, 0x6e, 0x69, 0x63, 0x3a, 0x20, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x20, 0x72, 0x65, 0x61, 0x64, 0x20, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x3a, 0x20, 0x76, 0x69, 0x65, 0x77, 0x2e, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x20, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x65, 0x73, 0x20, 0x6e, 0x6f, 0x20, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x20, 0x6f, 0x66, 0x20, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74, 0x3b, 0x20, 0x65, 0x78, 0x70, 0x65, 0x63, 0x74, 0x65, 0x64, 0x20, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74, 0x2c, 0x20, 0x66, 0x6f, 0x75, 0x6e, 0x64, 0x20, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0xa}, exitCode:70}",
    "witness": true
  },
  {
    "test": "TestCheckedViewV2RepresentationMutants",
    "package": "internal/oracle",
    "file": "internal/oracle/checked_views_v2_migration_test.go:88",
    "seconds": 0.604,
    "oracle": "Handwritten physical representation outputs and sanitizer expectations; no external reference run",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: checked_views_v2_migration_test.go:120: old-layout read survived",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCheckedViewFlagDowncast",
      "TestCheckedViewInterfaces",
      "TestCheckedViewObjects",
      "TestCheckedViewOptionalReadBoundary",
      "TestCheckedViewV2ReadsAfterWrites",
      "TestCheckedViewV2ArrayArmBoundary",
      "TestCheckedViewV2WrongFamilyMutant",
      "TestCheckedViewV2RepresentationMutants",
      "TestCheckedViewV2MembershipMutant",
      "TestCheckedViewV2TupleIdentityMutant",
      "TestCheckedViewV2CallableProducerMutant",
      "TestCheckedViewV2MovedStage3Results",
      "TestCheckedViewObjectPrimitiveSource"
    ],
    "evidence": "ADAMIC_MUTANT=W1 ADAMIC_BUILD_CACHE_DIR=/tmp/u057/cache/W1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestCheckedViewFlagDowncast|TestCheckedViewInterfaces|TestCheckedViewObjects|TestCheckedViewV2WrongFamilyMutant|TestCheckedViewV2RepresentationMutants|TestCheckedViewV2MembershipMutant|TestCheckedViewV2TupleIdentityMutant|TestCheckedViewV2CallableProducerMutant)$' > W1.log 2>&1; checked_views_v2_migration_test.go:120: old-layout read survived",
    "witness": true
  },
  {
    "test": "TestCheckedViewV2MembershipMutant",
    "package": "internal/oracle",
    "file": "internal/oracle/checked_views_v2_migration_test.go:159",
    "seconds": 0.441,
    "oracle": "Node source execution plus handwritten boundary/panic text and sanitizer/leak checks; full output comparison",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: checked_views_v2_migration_test.go:184: membership check mutant survived native: oracle.run{stdout:[]uint8{0x74, 0x72, 0x75, 0x65, 0xa}, stderr:[]uint8{}, exitCode:0}",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCheckedViewFlagDowncast",
      "TestCheckedViewInterfaces",
      "TestCheckedViewObjects",
      "TestCheckedViewOptionalReadBoundary",
      "TestCheckedViewV2ReadsAfterWrites",
      "TestCheckedViewV2ArrayArmBoundary",
      "TestCheckedViewV2WrongFamilyMutant",
      "TestCheckedViewV2RepresentationMutants",
      "TestCheckedViewV2MembershipMutant",
      "TestCheckedViewV2TupleIdentityMutant",
      "TestCheckedViewV2CallableProducerMutant",
      "TestCheckedViewV2MovedStage3Results",
      "TestCheckedViewObjectPrimitiveSource"
    ],
    "evidence": "ADAMIC_MUTANT=W1 ADAMIC_BUILD_CACHE_DIR=/tmp/u057/cache/W1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestCheckedViewFlagDowncast|TestCheckedViewInterfaces|TestCheckedViewObjects|TestCheckedViewV2WrongFamilyMutant|TestCheckedViewV2RepresentationMutants|TestCheckedViewV2MembershipMutant|TestCheckedViewV2TupleIdentityMutant|TestCheckedViewV2CallableProducerMutant)$' > W1.log 2>&1; checked_views_v2_migration_test.go:184: membership check mutant survived native: oracle.run{stdout:[]uint8{0x74, 0x72, 0x75, 0x65, 0xa}, stderr:[]uint8{}, exitCode:0}",
    "witness": true
  },
  {
    "test": "TestCheckedViewV2TupleIdentityMutant",
    "package": "internal/oracle",
    "file": "internal/oracle/checked_views_v2_migration_test.go:190",
    "seconds": 0.4,
    "oracle": "Node source execution plus handwritten boundary/panic text and sanitizer/leak checks; full output comparison",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCheckedViewFlagDowncast",
      "TestCheckedViewInterfaces",
      "TestCheckedViewObjects",
      "TestCheckedViewOptionalReadBoundary",
      "TestCheckedViewV2ReadsAfterWrites",
      "TestCheckedViewV2ArrayArmBoundary",
      "TestCheckedViewV2WrongFamilyMutant",
      "TestCheckedViewV2RepresentationMutants",
      "TestCheckedViewV2MembershipMutant",
      "TestCheckedViewV2TupleIdentityMutant",
      "TestCheckedViewV2CallableProducerMutant",
      "TestCheckedViewV2MovedStage3Results",
      "TestCheckedViewObjectPrimitiveSource"
    ],
    "evidence": "ADAMIC_MUTANT=W1 ADAMIC_BUILD_CACHE_DIR=/tmp/u057/cache/W1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestCheckedViewFlagDowncast|TestCheckedViewInterfaces|TestCheckedViewObjects|TestCheckedViewV2WrongFamilyMutant|TestCheckedViewV2RepresentationMutants|TestCheckedViewV2MembershipMutant|TestCheckedViewV2TupleIdentityMutant|TestCheckedViewV2CallableProducerMutant)$' > W1.log 2>&1; PASS observed",
    "witness": true
  },
  {
    "test": "TestCheckedViewV2CallableProducerMutant",
    "package": "internal/oracle",
    "file": "internal/oracle/checked_views_v2_migration_test.go:221",
    "seconds": 0.53,
    "oracle": "Node source execution plus handwritten boundary/panic text and sanitizer/leak checks; full output comparison",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestCheckedViewFlagDowncast",
      "TestCheckedViewInterfaces",
      "TestCheckedViewObjects",
      "TestCheckedViewOptionalReadBoundary",
      "TestCheckedViewV2ReadsAfterWrites",
      "TestCheckedViewV2ArrayArmBoundary",
      "TestCheckedViewV2WrongFamilyMutant",
      "TestCheckedViewV2RepresentationMutants",
      "TestCheckedViewV2MembershipMutant",
      "TestCheckedViewV2TupleIdentityMutant",
      "TestCheckedViewV2CallableProducerMutant",
      "TestCheckedViewV2MovedStage3Results",
      "TestCheckedViewObjectPrimitiveSource"
    ],
    "evidence": "ADAMIC_MUTANT=W1 ADAMIC_BUILD_CACHE_DIR=/tmp/u057/cache/W1 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestCheckedViewFlagDowncast|TestCheckedViewInterfaces|TestCheckedViewObjects|TestCheckedViewV2WrongFamilyMutant|TestCheckedViewV2RepresentationMutants|TestCheckedViewV2MembershipMutant|TestCheckedViewV2TupleIdentityMutant|TestCheckedViewV2CallableProducerMutant)$' > W1.log 2>&1; PASS observed",
    "witness": true
  },
  {
    "test": "TestCheckedViewV2MovedStage3Results",
    "package": "internal/oracle",
    "file": "internal/oracle/checked_views_v2_migration_test.go:252",
    "seconds": 0.693,
    "oracle": "Node source execution compared with native and JavaScript outputs, plus sanitizer/leak checks",
    "oracle_kind": "external-run",
    "kills": [
      "M3"
    ],
    "unique_kills": [],
    "last_proven_fail": "M3: checked_views_v2_migration_test.go:273: native: exit codes differ; oracle.run{stdout:[]uint8{}, stderr:[]uint8{0x61, 0x64, 0x61, 0x6d, 0x69, 0x63, 0x3a, 0x20, 0x70, 0x61, 0x6e, 0x69, 0x63, 0x3a, 0x20, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x20, 0x72, 0x65, 0x61, 0x64, 0x20, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x3a, 0x20, 0x73, 0x79, 0x6d, 0x62, 0x6f, 0x6c, 0x2e, 0x66, 0x6c, 0x61, 0x67, 0x73, 0x20, 0x69, 0x73, 0x20, 0x6e, 0x6f, 0x74, 0x20, 0x61, 0x20, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x3b, 0x20, 0x65, 0x78, 0x70, 0x65, 0x63, 0x74, 0x65, 0x64, 0x20, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x2c, 0x20, 0x66, 0x6f, 0x75, 0x6e, 0x64, 0x20, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0xa}, exitCode:70}",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestCheckedViewV2ReadsAfterWrites"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "PLower",
      "PC",
      "PJS"
    ],
    "subsumer_seconds": 0.852,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestCheckedViewFlagDowncast",
      "TestCheckedViewInterfaces",
      "TestCheckedViewObjects",
      "TestCheckedViewOptionalReadBoundary",
      "TestCheckedViewV2ReadsAfterWrites",
      "TestCheckedViewV2ArrayArmBoundary",
      "TestCheckedViewV2WrongFamilyMutant",
      "TestCheckedViewV2RepresentationMutants",
      "TestCheckedViewV2MembershipMutant",
      "TestCheckedViewV2TupleIdentityMutant",
      "TestCheckedViewV2CallableProducerMutant",
      "TestCheckedViewV2MovedStage3Results",
      "TestCheckedViewObjectPrimitiveSource"
    ],
    "evidence": "ADAMIC_MUTANT=M3 ADAMIC_BUILD_CACHE_DIR=/tmp/u057/cache/M3 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestCheckedViewFlagDowncast|TestCheckedViewInterfaces|TestCheckedViewObjects|TestCheckedViewOptionalReadBoundary|TestCheckedViewV2ReadsAfterWrites|TestCheckedViewV2ArrayArmBoundary|TestCheckedViewV2WrongFamilyMutant|TestCheckedViewV2RepresentationMutants|TestCheckedViewV2MembershipMutant|TestCheckedViewV2TupleIdentityMutant|TestCheckedViewV2CallableProducerMutant|TestCheckedViewV2MovedStage3Results|TestCheckedViewObjectPrimitiveSource)$' > M3.log 2>&1; checked_views_v2_migration_test.go:273: native: exit codes differ; oracle.run{stdout:[]uint8{}, stderr:[]uint8{0x61, 0x64, 0x61, 0x6d, 0x69, 0x63, 0x3a, 0x20, 0x70, 0x61, 0x6e, 0x69, 0x63, 0x3a, 0x20, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x20, 0x72, 0x65, 0x61, 0x64, 0x20, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x3a, 0x20, 0x73, 0x79, 0x6d, 0x62, 0x6f, 0x6c, 0x2e, 0x66, 0x6c, 0x61, 0x67, 0x73, 0x20, 0x69, 0x73, 0x20, 0x6e, 0x6f, 0x74, 0x20, 0x61, 0x20, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x3b, 0x20, 0x65, 0x78, 0x70, 0x65, 0x63, 0x74, 0x65, 0x64, 0x20, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x2c, 0x20, 0x66, 0x6f, 0x75, 0x6e, 0x64, 0x20, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0xa}, exitCode:70}",
    "witness": false
  },
  {
    "test": "TestCheckedViewObjectPrimitiveSource",
    "package": "internal/oracle",
    "file": "internal/oracle/checked_views_v2_object_primitive_test.go:14",
    "seconds": 1.604,
    "oracle": "Node source execution plus handwritten boundary/panic text and sanitizer/leak checks; full output comparison",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M2",
      "M3"
    ],
    "unique_kills": [],
    "last_proven_fail": "M3: checked_views_v2_object_primitive_test.go:72: exit codes differ: oracle.run{stdout:[]uint8{0x74, 0x72, 0x75, 0x65, 0xa}, stderr:[]uint8{0x61, 0x64, 0x61, 0x6d, 0x69, 0x63, 0x3a, 0x20, 0x70, 0x61, 0x6e, 0x69, 0x63, 0x3a, 0x20, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x20, 0x72, 0x65, 0x61, 0x64, 0x20, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x3a, 0x20, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x2e, 0x6b, 0x69, 0x6e, 0x64, 0x20, 0x69, 0x73, 0x20, 0x6e, 0x6f, 0x74, 0x20, 0x61, 0x20, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x3b, 0x20, 0x65, 0x78, 0x70, 0x65, 0x63, 0x74, 0x65, 0x64, 0x20, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x2c, 0x20, 0x66, 0x6f, 0x75, 0x6e, 0x64, 0x20, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0xa}, exitCode:70}",
    "verdict": "overlapping",
    "subsumed_by": [
      "TestCheckedViewOptionalReadBoundary",
      "TestCheckedViewV2MovedStage3Results"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "PLower",
      "PC",
      "PJS"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestCheckedViewFlagDowncast",
      "TestCheckedViewInterfaces",
      "TestCheckedViewObjects",
      "TestCheckedViewOptionalReadBoundary",
      "TestCheckedViewV2ReadsAfterWrites",
      "TestCheckedViewV2ArrayArmBoundary",
      "TestCheckedViewV2WrongFamilyMutant",
      "TestCheckedViewV2RepresentationMutants",
      "TestCheckedViewV2MembershipMutant",
      "TestCheckedViewV2TupleIdentityMutant",
      "TestCheckedViewV2CallableProducerMutant",
      "TestCheckedViewV2MovedStage3Results",
      "TestCheckedViewObjectPrimitiveSource"
    ],
    "evidence": "ADAMIC_MUTANT=M3 ADAMIC_BUILD_CACHE_DIR=/tmp/u057/cache/M3 timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '^(TestCheckedViewFlagDowncast|TestCheckedViewInterfaces|TestCheckedViewObjects|TestCheckedViewOptionalReadBoundary|TestCheckedViewV2ReadsAfterWrites|TestCheckedViewV2ArrayArmBoundary|TestCheckedViewV2WrongFamilyMutant|TestCheckedViewV2RepresentationMutants|TestCheckedViewV2MembershipMutant|TestCheckedViewV2TupleIdentityMutant|TestCheckedViewV2CallableProducerMutant|TestCheckedViewV2MovedStage3Results|TestCheckedViewObjectPrimitiveSource)$' > M3.log 2>&1; checked_views_v2_object_primitive_test.go:72: exit codes differ: oracle.run{stdout:[]uint8{0x74, 0x72, 0x75, 0x65, 0xa}, stderr:[]uint8{0x61, 0x64, 0x61, 0x6d, 0x69, 0x63, 0x3a, 0x20, 0x70, 0x61, 0x6e, 0x69, 0x63, 0x3a, 0x20, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x20, 0x72, 0x65, 0x61, 0x64, 0x20, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x3a, 0x20, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x2e, 0x6b, 0x69, 0x6e, 0x64, 0x20, 0x69, 0x73, 0x20, 0x6e, 0x6f, 0x74, 0x20, 0x61, 0x20, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x3b, 0x20, 0x65, 0x78, 0x70, 0x65, 0x63, 0x74, 0x65, 0x64, 0x20, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x2c, 0x20, 0x66, 0x6f, 0x75, 0x6e, 0x64, 0x20, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0xa}, exitCode:70}",
    "witness": false,
    "vacuous_subcases": "PC/PJS preserve lowering-refusal cases which do not reach those entries; no positive executed case survived its Lower probe."
  }
]
```

| ID | Origin file:line | Change | Observed failing rows |
|---|---|---|---|
| M1 | internal/lower/view_lazy.go:196 | invert demand for unsupported checked-view reads | none |
| M2 | internal/lower/interface_cast.go:70,392 | change canonical optional-read boundary wording in both paths | TestCheckedViewObjectPrimitiveSource, TestCheckedViewOptionalReadBoundary |
| M3 | internal/native/view_fields.go:31 | require boolean physical storage for ordinary checked fields | TestCheckedViewV2TupleIdentityMutant, TestCheckedViewFlagDowncast, TestCheckedViewObjects, TestCheckedViewV2ReadsAfterWrites, TestCheckedViewInterfaces, TestCheckedViewObjectPrimitiveSource, TestCheckedViewV2MovedStage3Results |
| M4 | internal/javascript/view_fields.go:22 | drop literal restrictions from ordinary checked fields | none |

Raw failures above include witness precondition failures. matrix.json separately records the five production rows. W1 disables disagreement at internal/oracle/oracle_test.go:716. Six witnesses fail; tuple-identity and callable-producer still pass, so their verdict is untrue. Witness production failures never count as kills.

Survivors: see observation-clean.log, observation-M1.log and observation-M4.log and the saved observation source. M1 changes the helper's unsupported-demand result on synthetic IR; the public array fixture is refused earlier. M4 drops ordinary literal restriction checking; the independent admitted source changes a field after downcast. These observations are supplemental witnesses, not additional mutants or unit rows.

Brief ambiguities and costs:
- The supplied file commit is older than current origin/main. Scope was verified using the actual list at the recorded starting commit.
- Three mixed top-level rows contain both control runs and built-in mutants. They are classified as witnesses as whole rows, so production failures cannot decide their verdicts.
- The compiler instruction limits rebuilds to four mutants, while three per row would require 39. Four spread mutations were selected before failures were inspected.
- M2 is a diagnostic constant change in both canonical paths, not a behavioral runtime change. Subsumption is a hint over one observed mutant each, not a deletion recommendation.
- The package baseline exceeded its test-binary budget. All matrices ran exactly the 13 listed rows; package and repository uniqueness remain unknown.
- ObjectPrimitiveSource's comment-boolean, comment-good and comment-flags-wrong subcases hard-skip pending views-v3. No requested top-level row skipped. Outside-slice WASI and opt-in skips observed in baseline are recorded in baseline.log; that interrupted run is not a complete skip inventory.
- Coverage lists reached Go CUT functions. Runtime C calls were not traced. No runtime C or TypeScript oracle edits were used.
- Lower's empty-program probe can panic in emitters. Each row was isolated, so later rows were not falsely recorded as failures from a prior panic. Probe failures do not establish useful production discrimination.
- PC/PJS do not affect negative subcases already refused by lowering. They are not evidence of an executed positive case accepting empty output.
- All handwritten refusal and panic strings are self oracles, even when the same row also runs Node. The array boundary source checks only success exit, a weaker pin than full output.
- The two surviving witnesses need their own assertion that the comparison rejected a counterfactual. Their passing W1 runs are concrete evidence, not a production-mutant verdict.
- No other package suite or full repository gate was run. Native products used distinct caches for each compiler mutant. Cold matrix times include rebuild and test execution; exclusive build time is not separately measurable from these logs.

Timing: tools were warm, setup.sh skipped; nproc 5. npm-ci.log records dependency installation. timings.json records command wall durations, isolated own-binary medians are in rows.json, and each native rebuild matrix's wall time is in audit-runs.json. The whole-package baseline cooked at 90.091 seconds; bounded baseline's own line was 7.19 seconds. All standalone production and probe diffs and W1 passed go vet; final bounded baseline and final vet passed after restoration.

Concrete survivor witnesses: M1 clean checkLazyViewReads returns "probe:1: Adamic 0.1 refuses checked view read of field field with unsupported dictionary contract..."; M1 returns "<nil>". This is synthetic helper IR, not a proven source-language miscompile. M4 clean JavaScript produces empty stdout, field-read panic stderr, exit 70; M4 prints "other\n", empty stderr, exit 0. Node prints "other\n", exit 0 in both runs. Commands: apply the standalone diff, temporarily restore the saved observation bridge/test, then ADAMIC_BUILD_CACHE_DIR=/tmp/u057/cache/observation-ID timeout 120 go test -v -count=1 -timeout 90s ./internal/oracle/ -run '^TestU057Observation$' > observation-ID.log 2>&1. The observation source is not a unit test change or a counted mutant.

Reproduction: source /workspace/adamic-tools/env.sh. Apply one standalone .diff at the recorded starting commit before running its logged command. Matrix logs were produced using the saved selector source, with ADAMIC_MUTANT set to the id. W1 is a separate allowed harness edit, not a production selector id. u057-audit.py reconstructs the switch from menu.json and probes.json, vets standalone diffs, runs matrices/probes, and restores source. The full reached-function list is reached-functions.txt, generated before mutation selection.

Measured wall costs: npm reports 0.475 seconds install time; setup 0 seconds (warm tools), nproc 5. Coverage 20.047 seconds, 39 isolated timing commands 102.599 seconds total; selector binary build 13.526 seconds; M1 through M4 cold matrix/rebuild commands 9.682, 9.477, 9.849, 9.881 seconds respectively. Vet, witnesses, probes, build, matrices and final checks together consumed 108.127 command wall seconds, excluding supplemental observation recompiles. Supplemental observations were not individually wall-timed; own binary lines are 0.222, 0.204, 0.206 seconds. Human/tool overhead and initial baseline wall time were not separately totaled. No exclusive native rebuild timing was isolated from execution.
