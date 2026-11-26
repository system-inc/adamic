# Writable Node flags: the test's wrong versions

`test.cjs` prints 126 rows, each `ok` on this head. Each wrong version below
edits `adapt.cjs` and turns the named rows red while every other row stays
green. Run the test from the repository root, on `$APPLIED_TREE`, a full tree
produced by `stage3/apply.sh`, as the README's run line gives it:

```sh
NODE_PATH="$PWD/stage3/api/node_modules" node \
    stage3/adapt/99-seats-scout24-writable-node-flags/test.cjs "$APPLIED_TREE"
```

A row named `X` below is the pair `X before` and `X after`, one per
snapshot, unless one snapshot is named. A form's row also carries its file,
as in `as target before src/compiler/zzFlagsForm.ts`; the lists here drop it,
except for `wrong-binding-unpinned`, whose 14 rows are named whole.
`wrong-expected-rows.tsv` in the lane's round 17 audit holds every row.

## The two edit planning versions

These change what the adapter writes, so the first control sees the wrong
bytes:

1. `wrong-cast-kept`: skips the cast removal at `utilities.ts:971`, the
   `aggregateChildData` write. Red: `control first apply`.
2. `wrong-readonly-kept`: plans no removal of `readonly` from
   `readonly flags: NodeFlags;`. Red: `control first apply`.

## The seventeen check versions

3. `wrong-scope-five-files`: removes the program-wide checker walk, the keyed
   pins and the owner checks, keeping only the text check of the five edited
   files. Red: 124 rows, every refusal row on both snapshots, including both
   split-pin rows. Only `control first apply` and `control replay` stay green.
4. `wrong-roots-ts-only`: loads only `.ts` roots under `src/` and walks the
   roots, not the program's files. Red, 6 rows: `imported .cts writer`,
   `imported .mts writer`, `imported .tsx writer`.
5. `wrong-query-cast-forms`: checks a write only when its receiver is shaped
   like a `Mutable` cast, a `Mutable` parameter or a variable initialized from
   one or from a call. Red, 32 rows: `alias of alias`, `angle target`,
   `as target`, `assigned alias`, `call receiver`, `conditional receiver`,
   `imported .cts writer`, `imported .mts writer`, `imported .ts writer`,
   `imported .tsx writer`, `non-null initializer`, `non-null target`,
   `parenthesized target`, `property receiver`, `satisfies target`,
   `type alias`.
6. `wrong-assertion-steps-dropped`: stops treating `as`, `<T>` and `satisfies`
   around a write target as the same write. Red, 6 rows: `angle target`,
   `as target`, `satisfies target`.
7. `wrong-binding-unpinned`: matches a write through a fresh factory node to
   the first unclaimed binding pin when its statement matches, without the
   binding initializer or owner. Red, 14 rows, as Loom's J-tfcr970 ran it:
   - `angle target before src/compiler/zzFlagsForm.ts`
   - `angle target after src/compiler/zzFlagsForm.ts`
   - `as target before src/compiler/zzFlagsForm.ts`
   - `as target after src/compiler/zzFlagsForm.ts`
   - `ninth alias write before`
   - `ninth alias write after`
   - `non-null target before src/compiler/zzFlagsForm.ts`
   - `non-null target after src/compiler/zzFlagsForm.ts`
   - `owner copy under a second same-name ancestor before`
   - `owner copy under a second same-name ancestor after`
   - `parenthesized target before src/compiler/zzFlagsForm.ts`
   - `parenthesized target after src/compiler/zzFlagsForm.ts`
   - `satisfies target before src/compiler/zzFlagsForm.ts`
   - `satisfies target after src/compiler/zzFlagsForm.ts`
8. `wrong-constraints-unfollowed`: stops following base constraints on key
   and receiver types, so a type parameter, conditional, indexed or `keyof`
   type is never read as one that could be `"flags"` or hold a `Node` member.
   Red: 98 rows. The 28 green are both controls and the rows for alias
   initializer swaps, a direct write in another function, removed final alias
   and keyed writes, the ninth alias write, the owner text decoy at load, the
   split pin, a changed pinned alias write, the pinned write moved into a
   class field initializer, a function expression, a nested named function or
   an arrow, and the pinned write's changed condition and its diagnostic.
9. `wrong-key-assignability-dropped`: reads a nonliteral key as possibly
   `"flags"` only when it is `string`, `any` or the literal itself. Red, 4
   rows: `string-mapping key`, `template-literal key`.
10. `wrong-edges-skipped`: lets an unpinned keyed write, an unresolved target
    and an `any` receiver pass instead of refusing them. Red, 18 rows:
    `any receiver`, `generic key`, `keyed write duplicated`,
    `keyed write moved into an arrow`, `literal-typed key`, `string key`,
    `string-mapping key`, `template-literal key`, `unresolved name`.
11. `wrong-direct-text-only`: matches a direct pin by its statement text
    alone, not its function. Red, 4 rows: `direct write in another function`
    on both snapshots, `ninth alias write after`,
    `pinned alias write changed after`.
12. `wrong-pin-count-dropped`: drops the check that every pin was matched.
    Red, 4 rows: `final alias write removed`, `keyed write removed`.
13. `wrong-load-text-search`: finds a pinned direct statement in its owner's
    text by string search instead of by parsing. Red, 2 rows:
    `owner text decoy refused at load`.
14. `wrong-owner-named-only`: takes only a named function declaration as a
    write's owner, not any function-like node. Red, 2 rows:
    `keyed write moved into an arrow`.
15. `wrong-owner-bound-by-claims`: drops the walk that finds each owner by its
    key, so an owner is bound only by the writes that claim it. Red, 10 rows:
    `owner copy under a second same-name ancestor`,
    `owner moved into an inactive block`,
    `renamed ancestor with its old subtree in a block`, `same-name stand-in`,
    `second node with an owner's key`.
16. `wrong-owner-first-match`: drops the refusal when an owner's key names
    more than one node. Red, 2 rows: `second node with an owner's key`.
17. `wrong-owner-key-declarations-only`: keys an owner by its declaring
    modules, classes and functions alone, without each step's slot or the
    blocks between. Red, 6 rows: `owner copy under a second same-name
    ancestor`, `owner moved into an inactive block`,
    `renamed ancestor with its old subtree in a block`.
18. `wrong-owner-text-ignored`: drops the comparison of each owner's whole
    text with its pinned `-before` or `-after` text. Red, 16 rows:
    `pinned write after an early return`,
    `pinned write moved into a class static block`,
    `pinned write wrapped in a loop`, `pinned write wrapped in a try`,
    `pinned write wrapped in an if`, `pinned write's condition changed`,
    `pinned write's condition changed, diagnostic`,
    `statement swapped above a pinned write`.
19. `wrong-split-keys-dropped`: still refuses a pin whose writes fall in two
    nodes, but without the pinned key and the write's owner key in the
    message. Red, 2 rows: `pin's writes split between two nodes`.

## Retired

- `wrong-pin-text-unchecked`, retired in round 17 at Compiler's word: the
  owner-text check subsumes the pin text, so no row can go red. "A version
  that can't fail is a test that proves nothing." It dropped the statement
  text from a binding pin's match, and Loom's J-1ahvk4h saw all 126 rows
  green, as the table predicted.
- `wrong-owner-static-block-dropped`, retired in round 14: the separate
  owner-kind clause is gone, and `wrong-owner-text-ignored` exercises its
  rows.
- `wrong-owner-field-initializer-dropped`, retired in round 14: the separate
  clause is gone. Its round 13 rows still refused because the initializer
  lacks expression-statement pin text.
- `wrong-path-unchecked` and `wrong-path-condition-text-dropped`, never
  written: the write path serves diagnostics, and the guard uses owner
  identity and text.

## Where they ran

Loom ran all 20, the 19 above and `wrong-pin-text-unchecked`, on head
16910bea at main b070eda2, each on a clean copy of the head against the
clean head's applied tree: J-zzsrdjm (versions 1 to 5), J-tfcr970 (6 to 8, 10
and 11), J-1ahvk4h (9, 12 to 14 and `wrong-pin-text-unchecked`) and
J-ge760sv (15 to 19). Each control was green at 126 rows. Every version
turned exactly its table's rows red but `wrong-binding-unpinned`, whose 14
rows are the ones listed above, where the lane's table had predicted 2.
