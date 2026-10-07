Optional-chain non-null assertions

This `.a` candidate follows Go chain membership and outermost-chain rules, skipping only parentheses when an assertion wraps a chain. Mid-chain assertions are declined. The diagnostic, suggestion ID, message, single operator removal, UTF-8 range, applied suggestion and unchanged automatic-fix result match the real Go rule. The semantic mutant removes the preceding byte instead of the assertion operator; it compiles, exits cleanly and is caught only by output comparison on Node, emitted JavaScript and sanitized native.

Complete suggestion objects use the owned model in the sibling no-non-null-assertion directory. That directory also owns the detailed comparison driver, Go output adapter, reproduction script and logs. This candidate refuses selection through the regular driver rather than losing suggestion edits. Shared `.a` registration and full suggestion reporting still require integration. No shared repository files were changed.
