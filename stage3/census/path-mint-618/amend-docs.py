from pathlib import Path
r=Path('/workspace/scratch/path-mint-618/work')
f=r/'stage3/adapt/99-correctness-timer-closures/README.md'
f.write_text(f.read_text()+'''
## Canonical Path result amendment

Adamic, #xn6fj0p, Nov 14, 2026. This amendment is type-only and follows
`typescript/scout-618` road (b): `toCanonicalFilePath` retains the string-returning
`createGetCanonicalFileName(useCaseSensitiveFileNames)` without a function cast.
The four Path-consuming calls mint their results with `as Path`: the directory
cache key, both recursive `invokeCallbacks` branches, and the child-watcher
cache lookup. The canonicalizer passed to `getRelativePathFromDirectory` and
the `includes` reader keep the string contract. No helper or normalization is
introduced. Only this owner's after text and SHA-256 change; the before pin
and the earlier timer and queue amendments retain their existing text.

The complete main/amendment single-file CommonJSBundle census delta, restoring
function-cast mutant, whole-owner contract, 79-root stock and Adamic checks,
per-file emitted JavaScript comparison, upstream `hereby dts` comparison, and
whole unfiltered lane receipts belong on `typescript/path-mint-618-evidence`,
under `stage3/census/path-mint-618/`.
''')
f=r/'stage3/upstream/LEDGER.md'
f.write_text(f.read_text()+'''\n- **Type-only, canonical Path result mints (#xn6fj0p):** 99-correctness-timer-closures amends createDirectoryWatcherSupportingRecursive to retain its string-returning canonicalizer without a function cast and apply `as Path` at exactly four Path-consuming calls, following typescript/scout-618 road (b); no runtime operation or normalization changes. Evidence: typescript/path-mint-618-evidence, stage3/census/path-mint-618/.\n''')
