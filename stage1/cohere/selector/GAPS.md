# CSS selector parser

A port of cohere's `internal/format/css/selector`, itself a port of
`postcss-selector-parser` 2.2.3. The supported entry is lossless default parsing,
which cohere's Prettier integration uses. This is a parser, not a selector matcher
or CSS printer. Regular expression splits are Adamic regular expression literals,
compiled natively on the regex-matcher base branch.

The AST is a flat arena of typed nodes, with child and parent indexes. Canonical
JSON preserves every own property, absent versus own undefined fields, source
locations, byte source indexes, raw attribute values, namespaces, spaces and
trailing commas. Parent pointers are excluded from comparisons. Go's UTF-16
index conversion happens inside `parse`, including its unusual positions within
surrogate pairs and beyond the input. Parse errors unwind with throw/catch.

`selector_test.go` compares byte output with Go cohere, source Node, the JavaScript
backend and, when `ADAMIC_SELECTOR_LIBRARY` is set, the original library. Every
native baseline runs under ASan and UBSan and receives a separate leak check.
Three mutated ports must finish successfully and disagree with Go on native and
Node. `gaps_test.go` checks each program below against its Node output and the
exact `lower.NotYet.What` refusal. A gap closing therefore fails the gap test.

## Compiler gaps

| Program                                                   | Node stdout | Exact refusal                                          | Workaround                                     |
| --------------------------------------------------------- | ----------- | ------------------------------------------------------ | ---------------------------------------------- |
| [1_multiple_push.ts](gaps/1_multiple_push.ts)             | `ab`        | `push with other than one value`                       | One push per value.                            |
| [2_mixed_field.ts](gaps/2_mixed_field.ts)                 | `true`      | `a field of type string \| boolean \| undefined`       | Namespace string and separate true bit.        |
| [4_array_from_iterator.ts](gaps/4_array_from_iterator.ts) | `a`         | `Array.from with other than { length } and a callback` | Explicit map-key collection loop.              |
| [5_conditional_panic.ts](gaps/5_conditional_panic.ts)     | `a`         | Closed by compiler/area-next: lowers, must match Node  | Guard helper calls panic in its own statement. |
| [6_undefined_case.ts](gaps/6_undefined_case.ts)           | `a`         | `a case whose type differs from the switch's`          | Narrow undefined before the switch.            |

Methods are ordered callee first, and parenthesis recursion is integrated into
`Parser.parse`, following the GraphQL slice's workaround for forward method
calls. That inherited limitation is not asserted as a new selector gap here.

compiler/area-next closed gap 5 and moved gap 6's refusal: `case undefined` in a
switch over `string | undefined` is still refused, now because the case's type
differs from the switch's. With gap 5 closed, `gaps_test.go` builds it natively and
requires Node's `a` with no leaks. The workarounds still stand; retiring gap 5's is
its own change against the corpus.

## Upstream nontermination

Go deliberately returns `ErrLoopsForever` for an unconsumed namespace bar;
Adamic returns the same message. The original library loops forever. The Node
comparison explicitly reports and excludes these inputs rather than claiming
agreement. `testdata/nontermination.mjs` is a proving program. Four subprocess
checks (`a| b`, `:is(a|)`, `a|@x`, `*|)`) require a marker written immediately
before calling the parser, then a deadline killing the process without a return
marker. Observation: each call failed to return within one second. Inference:
inspection of upstream's non-advancing parser dispatch explains the loop.

## Corpus and limits

The default seed is 20261005 and the generator produces 4,000 cases.
The exact cohere PCG(2,3) random oracle corpus adds 20,330 unique inputs:
20,000 terminating inputs plus 330 cases its upstream test skips. Every string
literal and concatenated string in CSS subtree tests is submitted, including
malformed inputs and nonselector test constants; selectors are also extracted
from valid stylesheet snippets using independent PostCSS. Every `.css` file in
cohere and its TypeScript submodule is scanned. This checkout has one CSS file
(a theme fixture without rules), and no CSS files in TypeScript. Consequently
its actual file scan contributes zero selectors; test snippets contribute 387
including repeats. Fixtures exercise all six attribute operators, flags, values,
escapes, comments, namespace forms, pseudos and nesting through depth 300.

The default corpus has 27,160 inputs. Go agrees with all three Adamic execution
paths on every input. Original JS agrees on 26,759 terminating inputs; 401 are
explicit upstream nontermination exclusions. Canonical dumps normalize NaN to
JSON null and unpaired surrogate code units to replacement characters, matching
Go's tree representation. This does not establish raw UTF-16 fidelity outside
that representation. Nondefault upstream options, mutation APIs, rendering,
arbitrary invalid UTF-8 and exhaustive grammar coverage are outside this slice.

## Boolean field gap retired

Gap 3 is removed from the refusal tests. SelectorNode.quoted now holds boolean | undefined directly; quotedPresent is gone. The port's own differential test covers both boolean values and absence in attribute nodes. The minimized program moved to internal/oracle/testdata/boolean_optional_selector.a, held to Node in both backends.
