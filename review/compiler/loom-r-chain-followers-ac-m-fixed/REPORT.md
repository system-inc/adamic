Built: the fully validated chain-followers recut as one commit on main.
Commits: source a79a5600420f46e3dba79d5e91ca7097232db102; base e1bd97c5034ddaac7c47938f587c304264443b7d; validated tree 751a01a57bd2cd8ab7bc1d319ccf20f7ecec8c02; delivery SHA in final response.
Commands and outputs: all eight touched package trees, reader/maps/parallel guards, a-check and vet passed; original logs are in ../loom-r-chain-followers-ac-m/.
Mutants: 16 semantic/IR/cache mutants and both header mutants caught; every measured leaf below 60 seconds.
Not covered: opt-in lanes listed in the detailed report, full repository gate or Loom submission.

Use compiler/loom-r-chain-followers-ac-m-fixed. The earlier compiler/loom-r-chain-followers-ac-m is superseded because I accidentally added its push logs as a second evidence-only commit. This replacement is a new branch with one parent on the same validated main base; no pushed history is rewritten. The compiler, tests, fixtures, status records and counts are byte-identical to the validated tree. Only this report directory is additional. This is the conservative naming assumption needed to satisfy one commit and no force push after that mistake.

The exact conflict resolution, all 15 counts rows, test commands/results, timing, disk-space recovery, original-source comparisons and runtime comparison are in [the detailed report](../loom-r-chain-followers-ac-m/REPORT.md). The conflict retains the branch's concrete callable/array destinations and main's cached impliedExpressionTarget fallback. No main infrastructure line was reverted. The required Path boundary expectation names the later ElementAccessExpression stop. No extra runtime C/H resolution line exists. Four existing a-check headers and all program bodies are retained. All three source stage3 status files are included and their complete package passed.

The commit retains the original source message, requested re-cut line, and Co-Authored-By: Ahra <ahra@ahra.ai> trailer. No PR, ahra invocation or Loom submission was made.
