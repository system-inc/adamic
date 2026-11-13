# One-file latent census — task 68x1t2y

**No: the five units no longer stop on the overload test at main tip; each has first lowering finding `NotYet` at `src/compiler/binder.ts:1109:17`, `writing a checked object field without its source-slot type certificate`. This is already true at ae342094’s parent.**

Compared `3045f23040c2d51679c428d6e89f99f8d150847c` (ae342094’s parent) with `6286add0122cb4fb0126384bfd6f5799b7a4fb0d` (origin/main, verified unchanged before evidence packaging). No feature merges or compiler changes.

The requested file is `src/compiler/transformers/destructuring.ts`; `src/compiler/destructuring.ts` does not exist. Its four assertion calls are at lines 601, 606, 611 and 616. The assertNode call at destructuring.ts:319 belongs to nested `emitBindingOrAssignment` (318:5), reached from `flattenDestructuringBinding` (244:1). In these fully adapted trees assertNode overload 1 is at debug.ts:281, rather than the historical line 319. Both final destructuring.ts and debug.ts are byte-identical across endpoints; their exact bytes and source manifests are retained.

`ae342094` adds the instance-mapper regression test; its parent already contains the rigid `typeMapper = nil` implementation. This comparison therefore does not isolate the implementation change.

| Original unit | Unit location in destructuring.ts | First lowering finding at both commits |
|---|---|---|
| flattenDestructuringBinding | 244:1 | NotYet binder.ts:1109:17 — writing a checked object field without its source-slot type certificate |
| makeArrayBindingPattern | 600:1 | NotYet binder.ts:1109:17 — writing a checked object field without its source-slot type certificate |
| makeArrayAssignmentPattern | 605:1 | NotYet binder.ts:1109:17 — writing a checked object field without its source-slot type certificate |
| makeObjectBindingPattern | 610:1 | NotYet binder.ts:1109:17 — writing a checked object field without its source-slot type certificate |
| makeObjectAssignmentPattern | 615:1 | NotYet binder.ts:1109:17 — writing a checked object field without its source-slot type certificate |

The nested emitBindingOrAssignment unit at 318:5 has the same first lowering finding. In full mode a “stop” is a failed construct boundary: the census rolls back and continues at the next statement/declaration. This is not proof that either assertion overload is now accepted; an earlier dependency finding can hide it.

| Totals | Parent | Tip |
|---|---:|---:|
| Catalogued declarations | 178 | 178 |
| Units, all attempted | 26 | 26 |
| Raw finding records | 89 | 89 |
| Raw Refused / NotYet / Boundary | 37 / 23 / 29 | 37 / 23 / 29 |
| Unique Refused / NotYet / Boundary, entire selected-file record | 36 / 11 / 28 | 36 / 11 / 28 |
| Unique Refused / NotYet / Boundary, sites attributed to destructuring.ts | 33 / 10 / 26 | 33 / 10 / 26 |
| Unique lowering Refused / NotYet / Boundary, including dependencies | 3 / 10 / 28 | 3 / 10 / 28 |

Unique totals deduplicate `(kind, where, reason, text)` across attempts and phases. Boundaries are recovery records, separate from compiler Refused/NotYet counts. The selected-file record includes dependency sites in other files; the attribution row counts only actual destructuring.ts sites.

**Units whose first finding changed: none.** First lowering findings also changed in no units. All findings, including order, match for every unit after replacing only the two absolute adapted-root prefixes with `<adapted>`. No units were added or removed.

The predicate selects six units: four make*Pattern units and flattenBindingOrAssignmentElement / flattenObjectBindingOrAssignmentPattern, which encounter `overload 1 of visitNode parameter node cannot be served by implementation parameter node` at visitorPublic.ts:123:5. No finding has a debug.ts site in 244–320. The original assertNode caller at 244:1 and its nested body at 318:5 are included below as two additional tracked units, even though neither now meets the predicate.

Both census invocations exited 0, took 55.673s and 57.256s, and have complete relation trailers with no overflow findings. Both headers have `checker_rejected: false`, empty diagnostics and diagnostic sites. The overlay’s historical fixed label still says “measured on a checker-rejected program”; that label is reproduced verbatim, not treated as the actual checker status.

Validation limitation: the bundled full_audit.py exited 1 at both commits. Its first assertion expects Refused reasons for string and number conditions plus var; actual reasons were only `{'var'}`. Its three probe executions completed, but the audit did not pass or validate its mutant assertions. The audit logs and probe streams are retained. No audit or compiler code was changed. `LATENT_ASSERT_NO_OUTPUT=1` was enabled in the real measurements, which exited successfully.

Each checkout ran its own stage3/apply.sh, make_overlay.py and go build. All six build/adapt/overlay steps exited 0. Every long command had a 1,700-second timeout; the wrapper prints, records and propagates its child exit status. Exact pins, product keys, binary/overlay/source hashes and settings are in [provenance.json](provenance.json). Raw streams, timing, commands, audit logs and exit codes are under parent/ and tip/. Compressed files retain original bytes.

Reproduction (substitute paths for each clean checkout and adapted output):

```sh
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/path/to/pinned/TypeScript-v6.0.3
timeout --preserve-status 1700 stage3/apply.sh /path/to/new-adapted
timeout --preserve-status 1700 python3 stage3/census/latent/make_overlay.py "$PWD" /path/to/new-overlay
gofmt -w /path/to/new-overlay/*.go
timeout --preserve-status 1700 go build -buildvcs=false -overlay=/path/to/new-overlay/overlay.json -o /path/to/census ./stage3/census/latent/tool
LATENT_FULL=1 LATENT_ONLY_FILE=/path/to/new-adapted/src/compiler/transformers/destructuring.ts LATENT_ASSERT_NO_OUTPUT=1 timeout --preserve-status 1700 /path/to/census /path/to/new-adapted/src/compiler /path/to/census.jsonl
```

The exact pinned parser packages must be installed in each checkout’s stage3/api; cohere and its TypeScript submodule must match the recorded pins. compare.py independently reads the two streams, checks full mode, complete trailers and exactly one selected file, computes the predicate union and totals, and compares findings with only root-path normalization. Run `python3 compare.py`, then `python3 render.py` to reproduce the derived comparison and this report.

## Verbatim findings at both commits

Every finding object for each selected/tracked unit follows, in stream order. No reason, text, location, phase or other finding field is normalized here; absolute paths remain exact. [comparison.json](comparison.json) also retains declaration metadata.

### flattenDestructuringBinding — destructuring.ts:244:1

parent (`3045f23040c2d51679c428d6e89f99f8d150847c`):

```json
[
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:244:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:277:180",
    "reason": "a value of type Expression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:277:180: Adamic 0.1 refuses a value of type Expression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace; keep pos readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:244:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:282:67",
    "reason": "a value of type ParameterDeclaration | VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:282:67: Adamic 0.1 refuses a value of type ParameterDeclaration | VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace; keep pos readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:244:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:294:17",
    "reason": "a value of type Expression[] | undefined seen as {}[] | undefined, which can write {} where Expression is read",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:294:17: Adamic 0.1 refuses a value of type Expression[] | undefined seen as {}[] | undefined, which can write {} where Expression is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:244:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:306:75",
    "reason": "a value of type Expression[] seen as {}[], which can write {} where Expression is read",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:306:75: Adamic 0.1 refuses a value of type Expression[] seen as {}[], which can write {} where Expression is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:244:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:309:22",
    "reason": "a value of type VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:309:22: Adamic 0.1 refuses a value of type VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace; keep pos readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:244:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/binder.ts:1109:17",
    "reason": "writing a checked object field without its source-slot type certificate",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:244:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:319:9",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet",
    "start": 13858,
    "end": 13908
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:244:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:308:9",
    "reason": "writing a checked object field without its source-slot type certificate",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:308:9: stage 0 can't lower writing a checked object field without its source-slot type certificate yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:244:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:269:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:308:9: stage 0 can't lower writing a checked object field without its source-slot type certificate yet",
    "start": 11180,
    "end": 12043
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:244:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:295:17",
    "reason": "a method sharing its name with a getter; rename the getter or use a callable arrow field in the interface and implementation",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:295:17: stage 0 can't lower a method sharing its name with a getter; rename the getter or use a callable arrow field in the interface and implementation yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:244:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:293:13",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:295:17: stage 0 can't lower a method sharing its name with a getter; rename the getter or use a callable arrow field in the interface and implementation yet",
    "start": 12667,
    "end": 12883
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:244:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:301:10",
    "reason": "a for...of destructuring an object",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:301:10: stage 0 can't lower a for...of destructuring an object yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:244:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:301:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:301:10: stage 0 can't lower a for...of destructuring an object yet",
    "start": 13029,
    "end": 13544
  }
]
```

tip (`6286add0122cb4fb0126384bfd6f5799b7a4fb0d`):

```json
[
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:244:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:277:180",
    "reason": "a value of type Expression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:277:180: Adamic 0.1 refuses a value of type Expression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace; keep pos readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:244:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:282:67",
    "reason": "a value of type ParameterDeclaration | VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:282:67: Adamic 0.1 refuses a value of type ParameterDeclaration | VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace; keep pos readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:244:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:294:17",
    "reason": "a value of type Expression[] | undefined seen as {}[] | undefined, which can write {} where Expression is read",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:294:17: Adamic 0.1 refuses a value of type Expression[] | undefined seen as {}[] | undefined, which can write {} where Expression is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:244:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:306:75",
    "reason": "a value of type Expression[] seen as {}[], which can write {} where Expression is read",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:306:75: Adamic 0.1 refuses a value of type Expression[] seen as {}[], which can write {} where Expression is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:244:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:309:22",
    "reason": "a value of type VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:309:22: Adamic 0.1 refuses a value of type VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace; keep pos readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:244:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/binder.ts:1109:17",
    "reason": "writing a checked object field without its source-slot type certificate",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:244:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:319:9",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet",
    "start": 13858,
    "end": 13908
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:244:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:308:9",
    "reason": "writing a checked object field without its source-slot type certificate",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:308:9: stage 0 can't lower writing a checked object field without its source-slot type certificate yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:244:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:269:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:308:9: stage 0 can't lower writing a checked object field without its source-slot type certificate yet",
    "start": 11180,
    "end": 12043
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:244:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:295:17",
    "reason": "a method sharing its name with a getter; rename the getter or use a callable arrow field in the interface and implementation",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:295:17: stage 0 can't lower a method sharing its name with a getter; rename the getter or use a callable arrow field in the interface and implementation yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:244:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:293:13",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:295:17: stage 0 can't lower a method sharing its name with a getter; rename the getter or use a callable arrow field in the interface and implementation yet",
    "start": 12667,
    "end": 12883
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:244:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:301:10",
    "reason": "a for...of destructuring an object",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:301:10: stage 0 can't lower a for...of destructuring an object yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:244:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:301:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:301:10: stage 0 can't lower a for...of destructuring an object yet",
    "start": 13029,
    "end": 13544
  }
]
```

### emitBindingOrAssignment — destructuring.ts:318:5

parent (`3045f23040c2d51679c428d6e89f99f8d150847c`):

```json
[
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:318:5",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:321:62",
    "reason": "a value of type Expression[] seen as {}[], which can write {} where Expression is read",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:321:62: Adamic 0.1 refuses a value of type Expression[] seen as {}[], which can write {} where Expression is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:318:5",
    "phase": "refusal_scan",
    "kind": "NotYet",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:324:36",
    "reason": "a read narrowed to undefined of a binding a call may write",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:324:36: stage 0 can't lower a read narrowed to undefined of a binding a call may write yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:318:5",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/binder.ts:1109:17",
    "reason": "writing a checked object field without its source-slot type certificate",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:318:5",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:319:9",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet",
    "start": 13858,
    "end": 13908
  }
]
```

tip (`6286add0122cb4fb0126384bfd6f5799b7a4fb0d`):

```json
[
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:318:5",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:321:62",
    "reason": "a value of type Expression[] seen as {}[], which can write {} where Expression is read",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:321:62: Adamic 0.1 refuses a value of type Expression[] seen as {}[], which can write {} where Expression is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:318:5",
    "phase": "refusal_scan",
    "kind": "NotYet",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:324:36",
    "reason": "a read narrowed to undefined of a binding a call may write",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:324:36: stage 0 can't lower a read narrowed to undefined of a binding a call may write yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:318:5",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/binder.ts:1109:17",
    "reason": "writing a checked object field without its source-slot type certificate",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:318:5",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:319:9",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet",
    "start": 13858,
    "end": 13908
  }
]
```

### flattenBindingOrAssignmentElement — destructuring.ts:338:1

parent (`3045f23040c2d51679c428d6e89f99f8d150847c`):

```json
[
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:338:1",
    "phase": "lowering",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/visitorPublic.ts:123:5",
    "reason": "overload 1 of visitNode parameter node cannot be served by implementation parameter node",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/visitorPublic.ts:123:5: Adamic 0.1 refuses overload 1 of visitNode parameter node cannot be served by implementation parameter node; make the implementation accept every value admitted by this overload, without mutable widening or bivariance"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:338:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:347:9",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/visitorPublic.ts:123:5: Adamic 0.1 refuses overload 1 of visitNode parameter node cannot be served by implementation parameter node; make the implementation accept every value admitted by this overload, without mutable widening or bivariance",
    "start": 14999,
    "end": 15130
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:338:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:348:13",
    "reason": "reading initializer",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:348:13: stage 0 can't lower reading initializer yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:338:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:348:9",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:348:13: stage 0 can't lower reading initializer yet",
    "start": 15130,
    "end": 16009
  }
]
```

tip (`6286add0122cb4fb0126384bfd6f5799b7a4fb0d`):

```json
[
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:338:1",
    "phase": "lowering",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/visitorPublic.ts:123:5",
    "reason": "overload 1 of visitNode parameter node cannot be served by implementation parameter node",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/visitorPublic.ts:123:5: Adamic 0.1 refuses overload 1 of visitNode parameter node cannot be served by implementation parameter node; make the implementation accept every value admitted by this overload, without mutable widening or bivariance"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:338:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:347:9",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/visitorPublic.ts:123:5: Adamic 0.1 refuses overload 1 of visitNode parameter node cannot be served by implementation parameter node; make the implementation accept every value admitted by this overload, without mutable widening or bivariance",
    "start": 14999,
    "end": 15130
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:338:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:348:13",
    "reason": "reading initializer",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:348:13: stage 0 can't lower reading initializer yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:338:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:348:9",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:348:13: stage 0 can't lower reading initializer yet",
    "start": 15130,
    "end": 16009
  }
]
```

### flattenObjectBindingOrAssignmentPattern — destructuring.ts:386:1

parent (`3045f23040c2d51679c428d6e89f99f8d150847c`):

```json
[
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:386:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:409:42",
    "reason": "a value of type BindingOrAssignmentElement[] | undefined seen as {}[] | undefined, which can write {} where BindingOrAssignmentElement is read",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:409:42: Adamic 0.1 refuses a value of type BindingOrAssignmentElement[] | undefined seen as {}[] | undefined, which can write {} where BindingOrAssignmentElement is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:386:1",
    "phase": "refusal_scan",
    "kind": "NotYet",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/binder.ts:1109:17",
    "reason": "writing a checked object field without its source-slot type certificate",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:386:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:420:99",
    "reason": "a value of type BindingOrAssignmentElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:420:99: Adamic 0.1 refuses a value of type BindingOrAssignmentElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace; keep pos readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:386:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:428:133",
    "reason": "a value of type ObjectBindingOrAssignmentPattern seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:428:133: Adamic 0.1 refuses a value of type ObjectBindingOrAssignmentPattern seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace; keep pos readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:386:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:429:82",
    "reason": "a value of type BindingOrAssignmentElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:429:82: Adamic 0.1 refuses a value of type BindingOrAssignmentElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace; keep pos readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:386:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/binder.ts:1109:17",
    "reason": "writing a checked object field without its source-slot type certificate",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:386:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:394:9",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet",
    "start": 17290,
    "end": 17758
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:386:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:395:57",
    "reason": "reading reuseIdentifierExpressions",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:395:57: stage 0 can't lower reading reuseIdentifierExpressions yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:386:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:395:9",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:395:57: stage 0 can't lower reading reuseIdentifierExpressions yet",
    "start": 17758,
    "end": 17854
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:386:1",
    "phase": "lowering",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/visitorPublic.ts:123:5",
    "reason": "overload 1 of visitNode parameter node cannot be served by implementation parameter node",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/visitorPublic.ts:123:5: Adamic 0.1 refuses overload 1 of visitNode parameter node cannot be served by implementation parameter node; make the implementation accept every value admitted by this overload, without mutable widening or bivariance"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:386:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:409:17",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/visitorPublic.ts:123:5: Adamic 0.1 refuses overload 1 of visitNode parameter node cannot be served by implementation parameter node; make the implementation accept every value admitted by this overload, without mutable widening or bivariance",
    "start": 18689,
    "end": 18823
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:386:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:418:21",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet",
    "start": 19293,
    "end": 19439
  }
]
```

tip (`6286add0122cb4fb0126384bfd6f5799b7a4fb0d`):

```json
[
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:386:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:409:42",
    "reason": "a value of type BindingOrAssignmentElement[] | undefined seen as {}[] | undefined, which can write {} where BindingOrAssignmentElement is read",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:409:42: Adamic 0.1 refuses a value of type BindingOrAssignmentElement[] | undefined seen as {}[] | undefined, which can write {} where BindingOrAssignmentElement is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:386:1",
    "phase": "refusal_scan",
    "kind": "NotYet",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/binder.ts:1109:17",
    "reason": "writing a checked object field without its source-slot type certificate",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:386:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:420:99",
    "reason": "a value of type BindingOrAssignmentElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:420:99: Adamic 0.1 refuses a value of type BindingOrAssignmentElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace; keep pos readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:386:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:428:133",
    "reason": "a value of type ObjectBindingOrAssignmentPattern seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:428:133: Adamic 0.1 refuses a value of type ObjectBindingOrAssignmentPattern seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace; keep pos readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:386:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:429:82",
    "reason": "a value of type BindingOrAssignmentElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:429:82: Adamic 0.1 refuses a value of type BindingOrAssignmentElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace; keep pos readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:386:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/binder.ts:1109:17",
    "reason": "writing a checked object field without its source-slot type certificate",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:386:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:394:9",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet",
    "start": 17290,
    "end": 17758
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:386:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:395:57",
    "reason": "reading reuseIdentifierExpressions",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:395:57: stage 0 can't lower reading reuseIdentifierExpressions yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:386:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:395:9",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:395:57: stage 0 can't lower reading reuseIdentifierExpressions yet",
    "start": 17758,
    "end": 17854
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:386:1",
    "phase": "lowering",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/visitorPublic.ts:123:5",
    "reason": "overload 1 of visitNode parameter node cannot be served by implementation parameter node",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/visitorPublic.ts:123:5: Adamic 0.1 refuses overload 1 of visitNode parameter node cannot be served by implementation parameter node; make the implementation accept every value admitted by this overload, without mutable widening or bivariance"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:386:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:409:17",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/visitorPublic.ts:123:5: Adamic 0.1 refuses overload 1 of visitNode parameter node cannot be served by implementation parameter node; make the implementation accept every value admitted by this overload, without mutable widening or bivariance",
    "start": 18689,
    "end": 18823
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:386:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:418:21",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet",
    "start": 19293,
    "end": 19439
  }
]
```

### makeArrayBindingPattern — destructuring.ts:600:1

parent (`3045f23040c2d51679c428d6e89f99f8d150847c`):

```json
[
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:600:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/binder.ts:1109:17",
    "reason": "writing a checked object field without its source-slot type certificate",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:600:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:601:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet",
    "start": 29704,
    "end": 29764
  }
]
```

tip (`6286add0122cb4fb0126384bfd6f5799b7a4fb0d`):

```json
[
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:600:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/binder.ts:1109:17",
    "reason": "writing a checked object field without its source-slot type certificate",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:600:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:601:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet",
    "start": 29704,
    "end": 29764
  }
]
```

### makeArrayAssignmentPattern — destructuring.ts:605:1

parent (`3045f23040c2d51679c428d6e89f99f8d150847c`):

```json
[
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:605:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:607:63",
    "reason": "a method read as a value (convertToArrayAssignmentElement would lose its object, and this with it)",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:607:63: Adamic 0.1 refuses a method read as a value (convertToArrayAssignmentElement would lose its object, and this with it); call it in an arrow that keeps the object: (element) => its object.convertToArrayAssignmentElement(element) (unbound-method)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:605:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/binder.ts:1109:17",
    "reason": "writing a checked object field without its source-slot type certificate",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:605:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:606:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet",
    "start": 29927,
    "end": 29999
  }
]
```

tip (`6286add0122cb4fb0126384bfd6f5799b7a4fb0d`):

```json
[
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:605:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:607:63",
    "reason": "a method read as a value (convertToArrayAssignmentElement would lose its object, and this with it)",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:607:63: Adamic 0.1 refuses a method read as a value (convertToArrayAssignmentElement would lose its object, and this with it); call it in an arrow that keeps the object: (element) => its object.convertToArrayAssignmentElement(element) (unbound-method)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:605:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/binder.ts:1109:17",
    "reason": "writing a checked object field without its source-slot type certificate",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:605:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:606:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet",
    "start": 29927,
    "end": 29999
  }
]
```

### makeObjectBindingPattern — destructuring.ts:610:1

parent (`3045f23040c2d51679c428d6e89f99f8d150847c`):

```json
[
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:610:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:612:47",
    "reason": "an unproven value assigned to a numeric literal or enum member slot SyntaxKind.BindingElement",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:612:47: Adamic 0.1 refuses an unproven value assigned to a numeric literal or enum member slot SyntaxKind.BindingElement; compare with this literal and return that constant, or widen the slot to the whole numeric enum or number (adamic/enum-literal)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:610:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/binder.ts:1109:17",
    "reason": "writing a checked object field without its source-slot type certificate",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:610:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:611:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet",
    "start": 30220,
    "end": 30275
  }
]
```

tip (`6286add0122cb4fb0126384bfd6f5799b7a4fb0d`):

```json
[
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:610:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:612:47",
    "reason": "an unproven value assigned to a numeric literal or enum member slot SyntaxKind.BindingElement",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:612:47: Adamic 0.1 refuses an unproven value assigned to a numeric literal or enum member slot SyntaxKind.BindingElement; compare with this literal and return that constant, or widen the slot to the whole numeric enum or number (adamic/enum-literal)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:610:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/binder.ts:1109:17",
    "reason": "writing a checked object field without its source-slot type certificate",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:610:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:611:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet",
    "start": 30220,
    "end": 30275
  }
]
```

### makeObjectAssignmentPattern — destructuring.ts:615:1

parent (`3045f23040c2d51679c428d6e89f99f8d150847c`):

```json
[
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:615:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:617:64",
    "reason": "a method read as a value (convertToObjectAssignmentElement would lose its object, and this with it)",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:617:64: Adamic 0.1 refuses a method read as a value (convertToObjectAssignmentElement would lose its object, and this with it); call it in an arrow that keeps the object: (element) => its object.convertToObjectAssignmentElement(element) (unbound-method)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:615:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/binder.ts:1109:17",
    "reason": "writing a checked object field without its source-slot type certificate",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:615:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-parent/src/compiler/transformers/destructuring.ts:616:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-parent/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet",
    "start": 30440,
    "end": 30513
  }
]
```

tip (`6286add0122cb4fb0126384bfd6f5799b7a4fb0d`):

```json
[
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:615:1",
    "phase": "refusal_scan",
    "kind": "Refused",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:617:64",
    "reason": "a method read as a value (convertToObjectAssignmentElement would lose its object, and this with it)",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:617:64: Adamic 0.1 refuses a method read as a value (convertToObjectAssignmentElement would lose its object, and this with it); call it in an arrow that keeps the object: (element) => its object.convertToObjectAssignmentElement(element) (unbound-method)"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:615:1",
    "phase": "lowering",
    "kind": "NotYet",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/binder.ts:1109:17",
    "reason": "writing a checked object field without its source-slot type certificate",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet"
  },
  {
    "root_kind": "",
    "depth": 0,
    "measurement": "measured on a checker-rejected program",
    "unit": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:615:1",
    "phase": "lowering",
    "kind": "Boundary",
    "where": "/tmp/census-68x1t2y/adapted-tip/src/compiler/transformers/destructuring.ts:616:5",
    "reason": "failed statement; state rolled back; continued at next statement or declaration",
    "text": "/tmp/census-68x1t2y/adapted-tip/src/compiler/binder.ts:1109:17: stage 0 can't lower writing a checked object field without its source-slot type certificate yet",
    "start": 30440,
    "end": 30513
  }
]
```
