# Single watcher forwarder evidence

Base: 10f8f6b4d0f7ce9d56f2558752eab3f095b6d9b2. Stock upstream: 050880ce59e30b356b686bd3144efe24f875ebc8 (6.0.3). Scout C is reproduced, including the local snapshot guard. No Loom submission.

| Check | Result |
|---|---|
| All-src count | 2 calls, 0 others; 3 total identifier uses across 712 source files, harness/testRunner included |
| Owner contract | 10 no-write refusals; LF/CRLF and unrelated additions accepted; second run 0 writes/edits, stable bytes/mtimes |
| Node witness | Both real nested caller paths match, using emitted full watch factory; registration before/during/after first event, removal during dispatch, identity close, cache reuse, length 3 and explicit undefined timestamp |
| Emission scope | Only helper, pollingWatchFile and fsWatch differ; all other sys.js bytes equal; only sys.ts changes among 747 source files |
| Adamic types | 79 roots together before/after; both exit 0, zero diagnostics |
| Stock tsc 6.0.3 | Before/adapted/mutant each 79 roots and 0 diagnostics; 0 new |
| Full single-file census | Full CommonJSBundle, no-output guard, 136 units in each tree; all 3 commands exit 0 |
| Census unique kinds | Refused/NotYet: 23/40 before, 22/41 adapted, 23/40 mutant; Boundary 55 throughout |
| Census raw | 190 before, 189 adapted, 190 mutant |
| Restoring mutant | sys-mutant.ts restores cast refusal at sys.ts:515:17; all source-aligned findings equal baseline |
| Adaptation 40 | Two historical declined any rows explicitly resolved at src/compiler/sys.ts:516:30 and :516:58, full source in resolved-sites.json; no executable reader of those historical dispositions found |
| Lane generated files | Sanction regeneration byte-identical; no hashed input moves; freshness check exit 0 |

## Complete census delta

The adapted exact tuple delta is 45 added and 45 gone, including every physical line shift. Full text and mapping are in adapted/findings-delta.json; adapted/findings-delta.md renders every row. The source-aligned delta contains exactly two new and two gone tuples, with one co-located Boundary change. There are no other source-aligned finding changes.

### Gone

- Boundary at `src/compiler/sys.ts:513:9`: failed statement; state rolled back; continued at next statement or declaration. Full diagnostic:

  ```text
  src/compiler/sys.ts:unmapped-base-514:17: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)
  ```

- Refused at `src/compiler/sys.ts:unmapped-base-514:17`: a cast the runtime can't check. Full diagnostic:

  ```text
  src/compiler/sys.ts:unmapped-base-514:17: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)
  ```

### New

- Boundary at `src/compiler/sys.ts:513:9`: failed statement; state rolled back; continued at next statement or declaration. Full diagnostic:

  ```text
  src/compiler/sys.ts:515:39: stage 0 can't lower a call through ?. (an optional call) yet
  ```

- NotYet at `src/compiler/sys.ts:515:39`: a call through ?. (an optional call). Full diagnostic:

  ```text
  src/compiler/sys.ts:515:39: stage 0 can't lower a call through ?. (an optional call) yet
  ```

The new optional call at 515:39 is #036ve6w’s; it remains deliberately unchanged. The mutant has 82 exact new/gone tuples from line shifts and zero source-aligned new/gone/changed findings. All streams retain the completed relation trailer, zero checker diagnostics, no usable IR, and the standard measurement-only label.

## Evidence scope and setup

The runtime VM exposes pollingWatchFile and fsWatch by extending only the emitted factory return object; actual helper, caller bodies and fsWatchHandlingExistenceOnHost execute unchanged. Workers inject events in place of OS delivery. This does not prove all OS error/fallback paths or native lowering.

SETUP.md discloses the initial default npm/temp/cache-location deviation. All later runs direct caches and outputs to scratch. Failed setup/comparison attempts are retained with their exits; neither compiler checks nor test expectations were weakened. See COMMANDS.md and commands.jsonl.gz for commands/exits.
