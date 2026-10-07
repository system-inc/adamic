Parked: react/boolean-prop-naming requires dynamic RegExp lowering, due Oct 9.
Reproducer: gaps/general_regex.a; see REQUIRED_FAILURE.md for the exact command and failure.
The private three-rule React driver also stops react-hooks/unsupported-syntax and react-hooks/use-memo before execution because it imports boolean_pattern.a. Their previous parity is historical, not current unified certification.
These ports are reserved on codex/typeaware-wave-03 and must not be copied into a landing branch until unblocked or isolated and certified.
The other twelve wave-03 ports pass their private checker suites but are not registered in the unified syntax harness. Running its tests on this branch does not certify them. A green unified landing branch cannot be claimed without migrating those ports and supplying their checker dependencies. No passing-only landing branch was pushed in this step.
