# Pending shared catalog adapter

The decision, message builder and live type query use the area RuleContext checker.
The factory refuses until RuleContext supplies the linked Go registry's React-value rule names.
No compiled-in list or private checker is used by this adapter.

Blocker: cohere/internal/lint/rules/structure/react_hook_no_any_type.go:290,
reactHookNoAnyTypeBlindedRules calls rule.Registered and filters ResolvesReactValueTypes.
RuleContext has no corresponding catalog helper. It is not a TypeScript checker question.
The upstream catalog mutant test adds a stand-in registration, so a frozen list cannot certify it.
Reproducer: go test ./internal/lint/rules/structure -run '^TestReactHookAnyTypeNamesTheBlindedRulesFromTheCatalog$' -count=1.

This directory is intentionally outside registry discovery. It is not green or registered.
