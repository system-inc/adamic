# Applied realpath proposal (historical patch)

`realpath.patch` records the original reviewed proposal, now applied with user
authorization. Do not apply it again. The implementation adds lexical input
normalization and Node fixtures beyond this historical diff. It adds a
minimal typed realPath filesystem primitive across declarations, IR, lowering,
freshness, C emission/runtime and the JavaScript oracle/backend. It changes no
internal/native/emit.go, internal/lower/lower.go, internal/native/native.go or
internal/oracle/oracle_test.go. The user explicitly authorized the supporting
compiler/runtime changes as a separate commit after review.

The patch uses realpath canonical keys when traversing project directories and
replaces TestDirectoryLinkGap with an exact Go comparison. An isolated workspace
with these edits passed:

```
go test -v -count=1 -timeout 30m ./stage1/cohere/config -run '^TestDirectoryLinksAgreeWithGoCohere$'
PASS: native ASan/UBSan, Node source and JavaScript backend emit a-alias/a.ts once.
ok github.com/system-inc/adamic/stage1/cohere/config 10.767s
```

Log: /tmp/stage1-config-realpath-proposal.log. This is a targeted alias test, not a
full runtime regression or filesystem-error validation. The applied continuation
adds direct Node canonical-path/error coverage, clean executable mutants and Go alias/cycle/dangling coverage, and runs compiler packages,
uncached input/count oracles and the configuration suite. See ../REPORT.md. The patch deliberately excludes the scratch
workspace's test helper adaptation for its symlinked cohere checkout.

Review with `git apply --stat stage1/cohere/config/testdata/realpath.patch`. The
patch uses zero context so storing it does not introduce whitespace errors from
context-line prefixes; application requires `git apply --unidiff-zero`. The historical patch is retained for review provenance; current files contain the completed authorized
continuation.
