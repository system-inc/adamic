# The scalar default value

Compiler accepts this type-only adaptation under Adamic's Nov 22 split (#bbssdcy, 05:38:48Z) after its review and census; the site's compiler road stays held as row 324 says, and this head doesn't touch it.

## The site

Tail family 80 (the tail pass at 26d185e5, project #bbssdcy): NotYet, "String
conversion of an object, array, map or function (ToPrimitive is not lowered)",
one unit, `generateOptionOutput`, at `src/compiler/executeCommandLine.ts:317:22`
in the adapted tree at 26d185e5 (upstream 6.0.3 `:317`). It is the argument of
`String(defaultValue)` in the fallback arm of the nested helper
`formatDefaultValue`. The helper's parameter is declared with the diagnostic
descriptor beside the scalars, so the conversion sees a typed mixed union and
stops.

## What reaches the helper

The helper, as upstream 6.0.3 declares it inside `generateOptionOutput`:

```ts
function formatDefaultValue(
    defaultValue: CommandLineOption["defaultValueDescription"],
    type: CommandLineOption["type"],
) {
    return defaultValue !== undefined && typeof type === "object"
        // e.g. ScriptTarget.ES2015 -> "es6/es2015"
        ? arrayFrom(type.entries())
            .filter(([, value]) => value === defaultValue)
            .map(([name]) => name)
            .join("/")
        : String(defaultValue);
}
```

By the time this adapter runs, 32-indexed-reads-program has given `arrayFrom` its type argument,
`arrayFrom<[string, string | number]>(type.entries())`, and the adapter's pin holds that text, the one line any
earlier adapter changes in the helper.

Its one call, the only other read of its name in the file:

```ts
const defaultValueDescription = typeof option.defaultValueDescription === "object"
    ? getDiagnosticText(option.defaultValueDescription)
    : formatDefaultValue(
        option.defaultValueDescription,
        option.type === "list" || option.type === "listOrElement" ? option.element.type : option.type,
    );
```

The first parameter is written `CommandLineOption["defaultValueDescription"]`.
Over the `CommandLineOption` union that is `string | number | boolean |
DiagnosticMessage | undefined`, the union of the field's declarations in
`types.ts` (`CommandLineOptionBase` and the string, number, boolean and custom
option kinds). The call sits in the false arm of `typeof ... === "object"`,
whose true arm already renders the descriptor through `getDiagnosticText`, so
its argument is `string | number | boolean | undefined` (triage-e/10's reading
of the checker). Inside the helper `defaultValue` is read three times: the
`!== undefined` test, the `=== defaultValue` comparison in the enum-name
lookup, and `String(defaultValue)`. None of them can meet a descriptor.

## The edit

One annotation: the first parameter's type becomes
`string | number | boolean | undefined`. The second parameter, the body,
`String(defaultValue)`, the call and every other byte stay as they are. The
annotation erases in emit, and the helper is nested and unexported, so the
emitted JavaScript and the public declarations are unchanged.

## What the adapter pins

Each pin is a node found through the syntax tree by its owner and kind, held
to its exact text, never a line number or a digest of the file:

- one `generateOptionOutput` implementation among the file's statements;
- one `formatDefaultValue` declaration among its body's statements, whose
  whole text (both parameters and the body, CRLF line endings) is the snapshot
  above, with the first parameter's type an indexed access type before and a
  union type after;
- the name `formatDefaultValue` read exactly twice in the file, its
  declaration and one call;
- that call as the false arm of the conditional quoted above, inside
  `generateOptionOutput`, passing `option.defaultValueDescription` first of two
  arguments.

A pin that is missing, duplicated or drifted refuses before anything is
written. A second run finds the adapted snapshot and writes nothing. The
adapter refuses unless the parser is stock TypeScript 6.0.3.

## Rulings

- Adamic's standing split (Nov 22, 05:38:48Z, #bbssdcy): a type-only
  adaptation where tsc stays clean, the JavaScript is identical, no public
  `.d.ts` line moves, and the census at main shows its sites cleared and 0 new
  is Compiler's accept.
- Row 324 (the String library row) and its Nov 22 correction keep this site's
  compiler road held: the row isn't amended for typed mixed unions. This
  adaptation converts no typed mixed union. It takes the descriptor out of a
  parameter no descriptor reaches.

## The proof

Type-only, so its proof is the emitted JavaScript byte for byte: stock tsc
clean over the adapted tree, each emitted `.js` and `.d.ts` of `src/compiler`
identical with and without this directory, the checker's resolved signature
of the call, the apply twice byte-identical, the restored annotation giving
the base's file, and the unit census at main showing the unit's stop cleared
with 0 units newly stopping.
