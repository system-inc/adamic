# Adjacent JSX recovery resolved

The previously blocked upstream case at `cohere/internal/lint/rules/react/jsx_no_duplicate_props_test.go:226`, `<A a /><B a />;`, now agrees with Go on source Node, emitted JavaScript and sanitized native after merging `origin/codex/parser-recovery-land` (`88f4a83d`). The case remains in the full 38-case capture inventory.
