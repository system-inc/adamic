# Runner seeds

These three tiny seeds hold the harness's three requested states on base main
`ef3d907`: NotYet, Refused, and Compiles with exact Node agreement.

- `01_queue_empty.a`: core.ts:1573's original nested `isEmpty` declaration and
  body, with a small wrapper supplying elements and headIndex. Empty, occupied,
  and fully consumed queue states are real tsc queue states. This stands for the
  census's 5,574 named nested declarations in 57 files. The enclosing createQueue
  mutation and dequeue logic are omitted to isolate the nested declaration.
- `02_is_string.a`: core.ts:1769's complete `isString` predicate unchanged,
  exercised on string and number. This stands for the 651 predicate nodes in
  31 files, not a claim that every predicate is verified.
- `03_return_true.a`: core.ts:1810's complete `returnTrue` unchanged, used as
  tsc's constant-true callback. This is the compiling control for boolean values
  used in conditions; it does not exercise the census's 6,697 non-boolean sites.

Node observations use the unmodified main `oracle/node.mjs`. All forms here are
erasable; no enum/namespace runner branch is needed. stage0.what retains the
complete compiler diagnostic with a repository-relative path, no CLI prefix or
trailing newline. These are runner seeds, not replacements for feature buckets;
other workers own their broader forms and hardest cases.
