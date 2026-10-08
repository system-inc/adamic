# Any-returning owners remain unfinished

Current main and the final type edits both have six observed function-return
blockers. Zero are adapted in this kind. Exact current sites:

- commandLineParser.ts:2437:10, convertConfigFileToObject.
- commandLineParser.ts:2467:17, convertToObject.
- commandLineParser.ts:2478:17, convertToObjectWorker.
- sys.ts:51:18, the host setTimeout declaration.
- utilities.ts:7809:17, tryParseJson.
- utilities.ts:8553:25, an objectAllocator constructor callback.

The JSON-producing owners need a recursive value union, including recovery's
undefined. The public input APIs also accept non-JSON JavaScript values; those
inputs must not be falsely narrowed to JSON. This is unfinished contract work,
not evidence that recursive JSON has no static type.

The timer owner must preserve each host's handle type through registration,
storage and cancellation. A Node-only return annotation would exclude browser
and custom host handles. The allocator returns an initially incomplete object
that becomes its advertised AST shape through later writes. Declaring the bare
constructor to already return a fully initialized AST would retain the lie.

Minimal contract programs for @system_adamic_typescript:

```a
const parsed = JSON.parse('{"value":"text"}');
console.log(parsed.value);
```

```a
class Header { kind = 0; }
const node = new Header();
// The actual object has no text until a later stage writes it.
```

Neither is claimed to pass Adamic. Decide the dynamic JSON boundary and staged
initialization contracts; do not infer impossibility from a checker error.

The small check control `function read(): number { return 1; }` compiles on
main. Restoring `: any` makes the same actual scratch program fail with
`a function returning any`, before a backend emits output. This mutant proves
the current refusal check; it does not prove a completed tsc owner adaptation.
No full-owner return mutant or return-kind oracle delta is claimed.
