# Wave 18 continuation stopped

Ahra corrected the automatic continuation request after claim commit `340ca63c`
was pushed. The original wave has two completed ports and one measured blocker.
The additional reservation is withdrawn from the current claim file; no
additional rules are claimed by this branch.

The outstanding original rule, `@next/next/no-title-in-document-head`, requires
JSX element parsing and nesting. The native parser lacks that grammar. The
unchanged Go rule produces a finding on the positive JSX control; native exits
70 with `parser slice expected GreaterThanToken, got Identifier` at offset 69.
The committed original report and logs contain this measurement. Completing
that port would require a parser change outside the rule's territory, so work
stops under Ahra's latest instruction. The shared harness and registration
generator were not edited.

The unfinished continuation draft is preserved locally under
`/workspace/wave-18-premature-continuation`, including 13 source/test files and
the two-line checker registration patch. It is not an integrated port. Its
initial comparison failed on a discarded-pure-result case; no agreement,
mutant or sanitizer success is claimed for the continuation. The shared
checker dispatcher has been restored to the previously pushed version.

The original implementation remains `adea6fbe`, with evidence commits
`fa3a1c1b` and `a7a173c0`. Its byte comparisons, rule/fact mutants, released
handle checks and sanitizers passed as recorded in WAVE_18_REPORT.md.
