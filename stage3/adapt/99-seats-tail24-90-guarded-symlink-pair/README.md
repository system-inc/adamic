Adamic accepts this JavaScript-changing adaptation after its review and census (#bbssdcy, the tail pass's family 90, rulings.md:24, :193 and :410).

Temporary under row 410. The original owner at the adapter base is below. The file-cache write precedes tuple selection, and the truthiness conditional is the final statement. The function has a void return and the new early return is reached only when the helper yields undefined. The local name `pair` is unused in this scope.

```ts
function processResolution(cache: SymlinkCache, resolution: ResolvedModuleFull | ResolvedTypeReferenceDirective | undefined) {
        if (!resolution || !resolution.originalPath || !resolution.resolvedFileName) return;
        const { resolvedFileName, originalPath } = resolution;
        cache.setSymlinkedFile(toPath(originalPath, cwd, getCanonicalFileName), resolvedFileName);
        const [commonResolved, commonOriginal] = guessDirectorySymlink(resolvedFileName, originalPath, cwd, getCanonicalFileName) || emptyArray;
        if (commonResolved && commonOriginal) {
            cache.setSymlinkedDirectory(
                commonOriginal,
                {
                    real: ensureTrailingDirectorySeparator(commonResolved),
                    realPath: ensureTrailingDirectorySeparator(toPath(commonResolved, cwd, getCanonicalFileName)),
                },
            );
        }
    }
```

The whole before and after function spellings pin the sole nested function named `processResolution` in `createSymlinkCache`. The adapter locates it through its syntax node and preserves line endings. Replay uses its own before and after snapshots. The witness compares calls, their arguments, and the return for an absent pair, a nonempty pair and every empty-string combination.

The support for the early return is `docs/rulings.md:208`: the original destructures the shared `emptyArray` when the helper yields undefined and the adapted function returns before any iteration, a difference only a program that writes `Symbol.iterator` on an array could observe, and row 208 refuses that write, so no admitted program tells the two apart.

- **Array membership in checked views waits for views v3; inferring physical storage from the asserted target is unsound; fixed object-backed tuples are admitted; an unused optional view is admitted lazily and reading its optional alias is refused.** NotYet and Refuse. Undated, compiler ruling (unattributed); source: docs/views-v2-dispatch-review.md:5-7,23,43-45. Spec: not in the spec. Test: internal/oracle/checked_views_v2_source_test.go:199 TestCheckedViewUntaggedArrayPending.
- **Conditions use JavaScript ToBoolean in .ts and .a (0, -0, NaN, '', null and undefined are false; represented objects and empty arrays are true), in `if`, loops, `?:` and `!`; optional numeric truthiness is direct ToBoolean; preferring explicit comparisons is taste for cohere:house; whether .a should require boolean conditions is undecided.** Accept. Oct 7 to 8 2026, @system_adamic; source: #bw3xg7c (5wqmh5cj), #yvgst44 (nz497e4d, item 8) / docs/verification/conditions/REPORT.md:10,24-26. Spec: What's in, Values and statements (conditions); The soundness rules (the stage 3 extension paragraph). Test: internal/oracle/conditions_test.go:25 TestTypeScriptConditionsAgreeWithNode.
- **Temporary adaptations to tsc's patch set are allowed when a language feature would land too late, tracked on the scoreboard and removed when it lands.** Meaning. Oct 7 2026, @system_adamic; source: #system_adamic_compiler_correctness (d5dfsn7f). Spec: not in the spec.
