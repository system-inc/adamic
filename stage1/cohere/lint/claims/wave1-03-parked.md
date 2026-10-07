# Parked rule

no-async-promise-executor is excluded from this landing branch. Its complete port and evidence remain at codex/lint-wave1-03 commit 3269b456a2e7b2cb808b9487f829bd43ed7232e8.

Reproducer: `new Promise(@dec async () => {})`. The unified TestRulesAgree captures this upstream recovery case; Adamic refuses unsupported primary AtToken at offset 12 before rule dispatch. This is the stage 1 parser gap filed with @system_adamic. No upstream case, shared comparator or parser is modified.

The other five owned rules remain registered: @typescript-eslint/no-unused-expressions, @typescript-eslint/unified-signatures, class-methods-use-this, no-case-declarations and no-compare-neg-zero.
