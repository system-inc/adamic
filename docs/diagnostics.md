# Proposed machine-readable diagnostics

This document proposes a JSON Lines interface for agents writing Adamic. It is **not implemented** by this unit. Today's human diagnostics remain the interface, and `Refused` remains distinct from `NotYet`: a language rule versus an implementation gap.

## One record per diagnostic

A future opt-in `--diagnostics=jsonl` mode should emit one UTF-8 JSON object per stderr line. Keep compiler output and artifacts off this stream. Emit the same first failure that human mode emits; this format must not change diagnostic ordering or accept/refuse outcomes. Never infer fields by splitting the human message or a colon-delimited path: construct them from source metadata and the rule's diagnostic producer.

Required fields:

| Field | Meaning |
|---|---|
| `version` | Schema version, initially `1`. |
| `kind` | `refused` or `not-yet`; checker and driver errors need separate kinds before they use this interface. |
| `location` | `{ "path": string, "line": integer or null, "column": integer or null }`. Lines and UTF-16 columns are one-based, matching `Program.Where`. Use the compiler's source path; do not invent a coordinate for filename-only diagnostics. |
| `rule` | Stable semantic ID independent of wording. Preserve existing IDs, including established lint IDs. Use `null` for an unmapped implementation gap until it has an assigned ID. |
| `construct` | The concrete failing construct or type, including contextual names. |
| `fix` | Actionable guidance, or `null` if no truthful alternative is known. A roadmap or reason alone is not a fix. |
| `replacement` | Suggested replacement object described below, or `null` when a local edit is unavailable. |

Optional `related` is an array of `{ "location": ..., "message": string }`, for example the write that can close a cycle. Optional `reason` explains the violated property without conflating it with the fix. Consumers must tolerate unknown optional fields; changing a field's meaning requires a new version. Rule IDs describe the restriction, not the Go construction site. A shared producer can select different IDs when its actual restrictions differ.

A replacement has `path`, `span: { start_byte, end_byte }`, `expected`, `text`, and `applicability: "review"`. Offsets are zero-based UTF-8 byte offsets in the original source file, and spans are half-open. The `expected` text must equal the source slice at that span; reject stale edits. A client should convert UTF-16 display positions using the original source text rather than treat them as byte offsets. Never fabricate an end position from today's `Where` string. The directive refusal currently uses a scanner helper that reports byte columns; a future producer should normalize it explicitly to UTF-16 rather than copying its human string. All proposals require review: even a syntactically supported replacement can change the author's intended behavior. Nonlocal fixes such as ownership redesign get `replacement: null`.

## Three worked examples

Each fenced line is one complete JSONL record, pretty context kept outside the record. These examples come from actual lower refusals. Example paths are normalized to make records reproducible; implementation should retain its supplied source paths.

### Coercing equality

For `equality.a` containing exactly `const same = 1 == 1;` plus a newline, `refusals.go` refuses the operator at line 1, column 16. The observed fix is “use ===, which doesn't coerce (adamic/strict-equality)”. The operator occupies bytes `[15,17)`.

```jsonl
{"version":1,"kind":"refused","location":{"path":"equality.a","line":1,"column":16},"rule":"adamic/strict-equality","construct":"==","fix":"use ===, which doesn't coerce","replacement":{"path":"equality.a","span":{"start_byte":15,"end_byte":17},"expected":"==","text":"===","applicability":"review"}}
```

### Function-scoped variable declaration

For `var.a` containing exactly `var old = 1;` plus a newline, `locals.go` refuses `var` at line 1, column 1. The observed fix is “use const or let (adamic/no-var)”. `let` is one suggested choice; the agent should consider whether the binding can be `const` and whether scope changes matter.

```jsonl
{"version":1,"kind":"refused","location":{"path":"var.a","line":1,"column":1},"rule":"adamic/no-var","construct":"var","fix":"use const or let","replacement":{"path":"var.a","span":{"start_byte":0,"end_byte":3},"expected":"var","text":"let","applicability":"review"}}
```

### A captured value that can close a reference cycle

The existing fixture `internal/oracle/testdata/fresh_refused/captured_before.a` creates `node`, creates `wrapper` with `nodes: [node]`, then writes `node.nodes.push(wrapper)`. The real diagnostic points to the `readonly nodes: Node[]` declaration at line 5, column 2 and names the potentially cyclic write at line 10, column 2. It offers weak elements, readonly arrays, or a proven fresh-write arrangement (`adamic/cycle-capable`). `readonly Node[]` can make the existing push illegal, while weak elements change read behavior, so no local replacement is claimed.

```jsonl
{"version":1,"kind":"refused","location":{"path":"internal/oracle/testdata/fresh_refused/captured_before.a","line":5,"column":2},"rule":"adamic/cycle-capable","construct":"Node[], an array whose elements can reach back to an array like it","reason":"A write may close a reference cycle that reference counting cannot free; the value written may reach what it is written into.","fix":"declare the elements weak, Weak<Node>[] (import type { Weak } from 'adamic'), which don't count and read undefined once what they point to is freed; or make it readonly Node[]; or write into such an array only values this function made, or only into one it made","replacement":null,"related":[{"location":{"path":"internal/oracle/testdata/fresh_refused/captured_before.a","line":10,"column":2},"message":"node.nodes.push(wrapper) may close the cycle"}]}
```

## A future implementation's checks

Keep human text and JSON fields derived from the same explicit diagnostic data, with stable IDs and fix text stored separately. Support unknown fixes and coarse locations honestly. Before implementing the mode, test JSON parsing, escaping paths and Unicode, one-based UTF-16 coordinates versus UTF-8 replacement offsets, stale-span rejection, null replacements, shared rule producers, and identical first-failure behavior across human and JSON modes. Compare accept/refuse probes across the implementation. The current wording audit and its remaining gaps are in [cloud/diagnostics-audit.md](../cloud/diagnostics-audit.md).
