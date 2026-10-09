u161 audited the complete two-test package at cf735d9fba9e38de6368575e5630e44375a86eaf.
Baseline green; no skips, missing rows, families, helpers or witnesses.
Six fixed-menu mutants: five killed, one survivor; both rows sacred within this package.
P1 empty main entry killed both rows; neither row is vacuous.
nproc 5; standalone diffs vetted and production source restored.

[
  {
    "test": "TestUnknownMutableContainerFailsLoudly",
    "package": "stage3/census/latent/statecopy",
    "file": "stage3/census/latent/statecopy/main_test.go:11",
    "seconds": 0.151,
    "oracle": "Self-written diagnostic substring, nonzero generator status and no published output. M3 and M6 demonstrate rejection of a different failure with the same nonzero status.",
    "oracle_kind": "self",
    "kills": [
      "M1",
      "M3",
      "M6"
    ],
    "unique_kills": [
      "M1"
    ],
    "last_proven_fail": "M1: main_test.go:34: unknown state shape survived: <nil>",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 6,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M1 timeout 120 go test -json -count=1 -timeout 90s ./stage3/census/latent/statecopy/ -run . > M1.log 2>&1; main_test.go:34: unknown state shape survived: <nil>",
    "timing_samples": [
      0.148,
      0.151,
      0.215
    ]
  },
  {
    "test": "TestUnknownForeignPointerFailsLoudly",
    "package": "stage3/census/latent/statecopy",
    "file": "stage3/census/latent/statecopy/main_test.go:41",
    "seconds": 0.165,
    "oracle": "Self-written diagnostic substring, nonzero generator status and no published output. M3 and M6 demonstrate rejection of a different failure with the same nonzero status.",
    "oracle_kind": "self",
    "kills": [
      "M2",
      "M3",
      "M5",
      "M6"
    ],
    "unique_kills": [
      "M2",
      "M5"
    ],
    "last_proven_fail": "M5: main_test.go:66: foreign mutable state survived: exit status 1 panic: state copy: foreign or unrecognized pointer  requires an explicit copier",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 6,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "evidence": "ADAMIC_MUTANT=M5 timeout 120 go test -json -count=1 -timeout 90s ./stage3/census/latent/statecopy/ -run . > M5.log 2>&1; main_test.go:66: foreign mutable state survived: exit status 1 panic: state copy: foreign or unrecognized pointer  requires an explicit copier",
    "timing_samples": [
      0.234,
      0.165,
      0.126
    ]
  }
]

| ID | Origin main.go line | Change | Failed rows |
|---|---:|---|---|
| M1 | 39 | named-container panic to return value | MutableContainer |
| M2 | 64 | unknown-pointer panic to return value | ForeignPointer |
| M3 | 90 | negate test-file exclusion | Both |
| M4 | 159 | output permission 0600 to 0644 | None |
| M5 | 27 | printed type string to empty string | ForeignPointer |
| M6 | 81 | argument count !=4 to ==4 | Both |

All locations are stage3/census/latent/statecopy/main.go on the starting commit. Short row names above denote the corresponding TestUnknown...FailsLoudly names. P1 is an empty main-entry return at line 80, separate from mutants and verdicts.

Survivor M4: with umask 000, a valid lowering struct generates identical compilable Go, but output mode changes 0600 to 0644. Commands and outputs are in survivor-controlled-umask.json. This is unguarded output-permission behavior. The initial witness used the inherited restrictive umask, which masked the change; survivor-witness.json retains that attempt.

Brief interpretation and costs:
- Both tests construct similar subprocess fixtures, but check different guards and diagnostics without a shared checker. They remain separate rows.
- Running go run builds the generator under test. It does not supply an external expected answer, so both oracles are self.
- A Go main function has no return value. P1 returns immediately, yielding success and no output; both tests detect this. Its standalone guard is always true but avoids unreachable-code vet errors.
- The four-mutant native rebuild guidance does not apply to this Go generator. Six fixed-menu mutants were declared before outcomes, spread over all three reached functions. No compiler or native port was mutated, so native build-cache isolation was unnecessary.
- Mandatory npm ci cost 1.424 seconds despite this package loading no Node modules. Warm toolchain setup cost zero.
- The default umask hid the permission mutation until the witness controlled it. This was measured, not treated as an equivalent mutation.

Scope and limits: complete package matrix, both rows observed for all six mutants and P1, no timeout or test-binary panic. Subprocess panics are expected and do not abort the matrix. No tests from other packages were run; repo-wide uniqueness remains for central replay. Other generator branches, successful copying semantics and generated-code alias behavior were not covered by the package's two negative tests or this mutant menu. All standalone diffs passed git apply --check and go vet; generated survivor witness outputs also passed go tool compile. Scratch switch is evidence only; production source was reverted.

Timings: baseline binary 0.295 seconds, wall 0.552 seconds. Solo binary samples: mutable 0.148/0.151/0.215 seconds, foreign 0.234/0.165/0.126 seconds. Aggregate measured operation wall times:
{
  "vet_seconds": 0.9176471449973178,
  "switch_build_seconds": 0.4624177810001129,
  "matrix_and_probe_wall_seconds": 2.911509116000161,
  "solo_wall_seconds": 4.989428840999608
}
