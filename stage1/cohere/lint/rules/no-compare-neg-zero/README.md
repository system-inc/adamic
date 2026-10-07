# Owned rule port

Rule code and messages are .a files. rule.json registers the concrete listener;
oracle.go selects the actual pinned Go rule. mutant.json names a compiling
semantic mutation, and testdata/witness.ts.txt is raw oracle input.

See ../no-async-promise-executor/REPORT.md for independent Go parity on source
Node, emitted JavaScript and sanitized native, compiler/stage1 coverage, complete
suggestion edits, mutant results, rates and explicit shared gaps. No shared
registration, test-harness, parser or compiler file is edited.
