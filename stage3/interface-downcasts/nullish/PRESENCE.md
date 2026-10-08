# Optional-number presence probe, October 7

Measured a detached checkout of codex/optional-field-write-2 at
86b3fe992c90969ae49588ff94116dd2307099c5. No code was changed there.
Built with `go build -buildvcs=false -o /tmp/nullish-presence/adamic ./cmd/adamic`.
The shared cohere submodule has the same pinned 7945d102 revision on both branches;
its implementation was referenced through a symlink, never copied.

After `record.first = undefined`, Node prints `true`, `[first]`, `true` for
`'first' in record`, Object.keys and Object.hasOwn. Keys are formatted with join
because this branch's console prelude accepts strings only.

The exact `first?: number` source is refused by the branch's checker, TS2412.
The checker-accepted `first?: number | undefined` control on an initially empty
interface value still cannot lower Object.hasOwn: its shape proof rejects the
interface annotation. The version observing in and Object.keys passes unchanged
against Node, sanitized native and JavaScript, all exit 0:

```text
false
[]
true
[first]
```

For all three observations, a plain-literal binding with `first: number | undefined`
is shared with an optional-number interface alias. Delete through the alias, then
write undefined through it. This uses the same physical slot and shared object.
Node and JavaScript print:

```text
false
[]
false
true
[first]
true
```

Native prints:

```text
false
[first]
false
true
[first]
true
```

The final three lines match: present-and-undefined is retained and observable as
present, distinct from absent. The first three reveal a separate enumeration bug:
deleting through the optional alias makes in and hasOwn false, but the literal's
Object.keys optimization still lists first. Thus the full absence/alias control
is not a green Node match; no broader representation completion is claimed.
No checker option was disabled, and no observation guard was bypassed.

Exact source fixtures and machine-readable command/exit/stdout/stderr rows are in
presence/. The initial CLI VCS stamping failure was avoided with buildvcs=false;
all final compiled runs used the configured clang and the branch's oracle/node.mjs.
The diagnostic split is delegated to TypeScript's ledger worker, per user steering.
