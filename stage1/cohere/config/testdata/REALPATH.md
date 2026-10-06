# Unapplied realpath proposal

`realpath.patch` is a reviewable continuation, not applied compiler code. It adds a
minimal typed realPath filesystem primitive across declarations, IR, lowering,
freshness, C emission/runtime and the JavaScript oracle/backend. It changes no
internal/native/emit.go, internal/lower/lower.go, internal/native/native.go or
internal/oracle/oracle_test.go. The unit's stage1-only territory is why applying the
supporting changes requires the scope approval already requested from the user.

The patch uses realpath canonical keys when traversing project directories and
replaces TestDirectoryLinkGap with an exact Go comparison. An isolated workspace
with these edits passed:

```
go test -v -count=1 -timeout 30m ./stage1/cohere/config -run '^TestDirectoryLinksAgreeWithGoCohere$'
PASS: native ASan/UBSan, Node source and JavaScript backend emit a-alias/a.ts once.
ok github.com/system-inc/adamic/stage1/cohere/config 10.767s
```

Log: /tmp/stage1-config-realpath-proposal.log. This is a targeted alias test, not a
full runtime regression or filesystem-error validation. Applying it should also
add canonical-path primitive error/loop coverage and run affected compiler tests,
the input oracle and the stage1 packages. The patch deliberately excludes the
scratch workspace's test helper adaptation for its symlinked cohere checkout.

Review with `git apply --stat stage1/cohere/config/testdata/realpath.patch`. The
patch uses zero context so storing it does not introduce whitespace errors from
context-line prefixes; application requires `git apply --unidiff-zero`. Its
existence does not authorize compiler changes.
