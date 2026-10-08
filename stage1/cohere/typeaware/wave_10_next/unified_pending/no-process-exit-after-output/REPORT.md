# Pending shared multi-file certification

The adapter delegates every current-file query to RuleContext.checker. It has no program opener.
The analysis, descriptions and repairs are copied from the previous native-certified port.
The directory is outside unified registry discovery and is not certified or registered.

Blocker: correctnessNoProcessExitAfterOutputIsProcessMember, cohere/internal/lint/rules/nexus/correctness_no_process_exit_after_output.go:410.
It needs declaration-file fixtures or resolved companion modules. The upstream fixture runner
is RunTypedFiles in cohere/internal/lint/testing/program.go:106. The capturedRun at line 313
stores len(files)-1, while docs_capture.go:53 records only the subject Source and an OtherFiles count.
The lint harness's shared_test.go:254 record shape and lines 296-303 replay only that subject.
It cannot replay this rule's upstream typed project or provide a firing Node witness with node.d.ts.
A subject containing console.log("output"); process.exit(0); loses its Process declarations
when replayed alone. It must not be certified from two matching zero-finding outputs.

The saved port's other-file asks additionally require a raw foreign-node question at the area checker
boundary. Checker.askFile currently anchors questions to this file's root. This question is not
implemented in these pending files; no private bridge path is used instead.

Reproducer in cohere: go test ./internal/lint/rules/nexus -run '^TestCorrectnessNoProcessExitAfterOutputFires$' -count=1 -v.
Upstream passing fixtures require the node.d.ts and other files in the test's files() / RunTypedFiles map.
The upstream reproduction log is saved in the landing evidence. Keep all fixture files and options
in the shared capture before activating the descriptor and asserting witness and mutant parity.
