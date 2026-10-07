Built: collapse.NodeStylesheetResolver, one .a file, including POSIX Go path joining and independent captured roots.
Commits: claim 4b887c85 pushed before code; implementation commit is this report's commit.
Commands: owned go test PASS 5.615s, 2,179 cases and 650,763 Go bytes in source Node/emitted JavaScript/sanitized native; owned go vet PASS, empty output.
Mutants: omit .css extension, broaden tailwindcss/ prefix, use importing base instead of captured package root; all compile and finish cleanly before comparison catches each on all three Adamic modes.
Not covered: Windows filepath semantics, whole-rule findings, disk access, full gate; six dependency occurrences removed, zero completely helper-ready rules.

Go cohere's unmodified NodeStylesheetResolver at 715ba94f3608a6500086b1076ce5cb7e51b836db supplies every expected path through an oracle-only virtual-main overlay. The submodule is unchanged. This function resolves strings without any disk access or Node invocation and always returns a nil error. Its Adamic API therefore returns the successful path directly. There are no external helper callbacks in this port.

The port reproduces POSIX filepath.Join, including non-resetting later absolute components, empty/empty join, dot segments, leading-root protection, unresolved relative .., repeated slashes and literal backslashes. It preserves the exact tailwindcss special case and tailwindcss/ boundary, strips that prefix, then appends .css only when the joined path lacks the exact lowercase suffix. Each returned function captures its factory's root independently.

Capture reads every Go string literal in matching test files from all six consumer families: canonical 204, class order 316, variant order 52, shorthand 178, conflicting 195, unknown 184. The 1,129 literals are supplemented by all combinations of 10 package roots, 7 importing bases and 15 specifiers, giving 2,179 cases. Each input observes a first factory, a distinct second factory, then the first again; expected paths are rendered as UTF-16 unit sequences, without sorting or guessed normalization. Consumer literals include source, options and expected strings; these are direct helper inputs rather than whole-rule finding replays.

Consumers:

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

The at-rule splitter, modifier parser and resolver together remove eighteen dependency occurrences across these six rules. Their remaining design-system, CSS and rule-local dependencies are not declared implemented. NewTheme is withdrawn to slot 08 and is not counted. These helpers produce no lint findings, so findings-per-second does not apply.

Reproduce with the setup environment sourced:

```
go test ./stage1/cohere/lint/helpers/from-wave1-13/resolver -count=1 -v -timeout=10m > /tmp/resolver.log 2>&1
go vet ./stage1/cohere/lint/helpers/from-wave1-13/resolver > /tmp/resolver-vet.log 2>&1
```

No compiler or shared harness was changed.
