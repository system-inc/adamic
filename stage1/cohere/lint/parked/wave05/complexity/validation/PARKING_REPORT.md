Parked the owned rule commits on origin/main f8013f0baac41ddc340d76f83bddde38536a8f07 under Ahra's shared-harness exception.
Owned code tip 6c3b953594270e53ece9924733d433da08e3b29c; the following report commit is the pushed parking tip. Helper branch a71683ad4 is already green on the same main.
Independent Go comparisons PASS on source Node, emitted JavaScript and sanitized native; six-rule aggregate PASS 265.827s, with 371 compiler/stage1 files and supported upstream corpora.
Every owned semantic mutant was caught on all three runtimes; 13 numeric listener mutants were also caught in 39 comparisons.
Pending: shared directory registration, context/Diagnostic and suggestion integration (#zmh9v36); JSX/parser and boundaries resolver limits remain explicit. No full repository gate or node-local dispatch certification.

Parking rebase used git rebase --onto origin/main 02ebc4e66 codex/lint-wave1-05. This excludes the old conflicting shared foundation history and retains this worker's owned rule and claim commits. Main's shared lint files remain unchanged. All 304 owned files were checked byte-identical against the isolated validation snapshot before testing. The previous pushed tip c7cecebb remains available in history and retained test support.

The temporary /tmp/w05-parking-repro snapshot contains current main's compiler and stage1 sources, the exact parked owned files, and test-support files from the former foundation. No legacy support is shipped onto main by this branch. Independent validators and the older scratch overlay use that support to compare rule bodies; this does not prove current main discovers these unregistered modules. The initial aggregate failed on boundaries' decision-only witness; a filtered replacement retained the six earlier rules while the later families used their independent validators. Initial symlinked Go sources violated Go internal-package path checks; a real pinned cohere source copy fixed that test setup. All final runs exited zero.

Commands, with source /workspace/adamic-tools/env.sh, in the isolated snapshot:
- python3 stage1/cohere/lint/rules/complexity/validation/validate.py
- python3 stage1/cohere/lint/rules/typescript-prefer-function-type/validation/validate.py
- python3 stage1/cohere/lint/rules/tailwind-no-unnecessary-whitespace/validation/validate.py
- python3 stage1/cohere/lint/rules/boundaries-dependencies/validate.py
- python3 stage1/cohere/lint/rules/next-google-font-display/validate.py
- python3 stage1/cohere/lint/rules/complexity/validation/validate_listeners.py
- ADAMIC_GATE_UNCACHED=1 go test -overlay=/tmp/w05-parking-second-overlay/overlay.json ./stage1/cohere/lint -run '^(TestWave05SupportedCorpus|TestWave05SecondCorpus|TestCompilerAndStage1Agree|TestMutants)$' -count=1 -v -timeout=20m

Complexity: 201 upstream cases, 351 compiler/stage1 files, 13,425,677 comparison bytes; modified-switch mutant caught. Function type/namespace/triple slash: 39/22/70 upstream cases, 351 files, 39,582,002 bytes; all three mutants caught. Whitespace: four supported upstream cases, 21 JSX exclusions, 351 files, 13,185,533 bytes; separator mutant caught. Boundaries: nine real decoded configurations, 7,742 decisions, 1,941,874 bytes; policy-order mutant caught, resolver findings remain blocked. Google font: 85,014 decision bytes; display mutant caught, JSX findings remain blocked. Listener metadata: 574 bytes, 13 rules, 25 entries; 39 mutation catches.

Earlier six-rule aggregate: 371 compiler/stage1 files, 15,221,259 bytes; 551 supported upstream cases, two JSX and one malformed comment exclusion; the second three-rule corpus is 100 cases without exclusions, 67,426 bytes. Owned assignment-token, RTL exemption, description, parenthesized-chain, postfix-assignment and compound-this-alias mutants passed their failure checks. Inherited mutants passed too.

Current throughput observations (native / Node / Go findings per second): complexity 965.99 / 1529.84 / 6726.55; function type 258.00 / 663.28 / 4998.85; namespace 973.99 / 1249.54 / 4916.14; triple slash 576.08 / 895.44 / 4013.96; whitespace 846.02 / 962.77 / 3670.89. These are best-of-three observations for each validator's compiler-plus-stress corpus. Correctness native is sanitized; throughput native is unsanitized. No general speed claim.

Numeric listener declarations remain metadata for the forthcoming driver. Existing rule bodies still depend on the old string-kind/index context and some subtree traversal. They must be adapted when the shared numeric-node driver and Diagnostic contract lands; no speed compliance is claimed from metadata alone. The parking blocker is missing shared registration/context/Diagnostic integration, owned by #zmh9v36. Rebase again onto the named harness sha before taking further work when it lands.
