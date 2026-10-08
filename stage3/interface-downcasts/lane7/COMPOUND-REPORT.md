Built finite disjoint tagged-arm runtime checks and verified original read spans.
Commit: see branch tip; parent 60d50afce8aa6e57dca9aabcc6596d9f4b4c653b.
Validation: checked-view gate, touched package filter, vet and source-arm mutants.
Mutants: skip selected fields, accept boolean text, drop nested fields each break three refusal pins.
Pending: full declarations, recursive runtime, untagged unions, arrays/callback compounds; 17 candidates / 66 reads remain.

Working date remains October 11, 2026, 23:00 UTC. Candidate counts are static,
not allocation-exact reachability. The previous 15 / 64 object queue gained two
builder callback overlaps / 2 reads. Five overlaps / 11 reads belong outside
this object lane; their brands are numeric with nonphantom declarations.

All 22 queued pairs / 77 reads have original upstream witness spans verified
against commit 050880ce59e30b356b686bd3144efe24f875ebc8. Metadata includes source
hashes and narrow context. Automatic approval review rejected staging complete
cohere files; direct git-object inspection and narrow witness extraction succeeded.
No upstream file was copied by this group.

The leading 11-read BindableStaticAccessExpression.expression pair now has three
finite tagged arm fixtures: identifier, property access and element access.
Wrong nested string shape, missing escapedText and unknown SyntaxKind each have
an exact exit-70 field refusal pin. Source Node prints true; sanitized native,
release native and generated JavaScript validate the same read and refuse wrong
values. These reduced declarations exclude full Identifier and Node members;
they are original-site-linked shape evidence, not certification of upstream pairs.
The tagged First & Second compound read is now checked; an untagged counterpart
still refuses at demand and unread casts remain admitted.

The new hook initially affected an existing plain-union diagnostic. The gate
caught it; eligibility is now restricted to intersection arms or previously
refused aliases. Recursive union selection stays refused to avoid incomplete
optional descriptor snapshots. Each selected arm reads the discriminator once.

Final gate: ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedView' -count=1 passed in 46.479s. Touched package filter passed; vet passed. Evidence is in compound-evidence.
