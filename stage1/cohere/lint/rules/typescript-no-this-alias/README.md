# TypeScript this-alias candidate

Own-directory .a implementation of `@typescript-eslint/no-this-alias`.
Filename gates, narrow initializer matching, assignment target unwrapping, compound
operators, allow lists, destructuring defaults, messages and ranges match Go on
the recorded corpus. Manifest field 5 carries decoded Go options, as specified
by the registry contract: ReportDestructuring and AllowedNames. The independent
adapter consumes that same decoded shape without changing the upstream rule.
Upstream configuration spellings and defaults are exercised through its captured
rule tests; arbitrary invalid configuration diagnostics are outside this check.

The default registry still needs .a discovery. validate.py uses a scratch Go
overlay and the adjacent Tailwind directory's unapplied compatibility.patch.
No shared registry, compiler, parser or oracle files were edited.

```
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/tmp/wave07-typescript python3 stage1/cohere/lint/rules/typescript-no-this-alias/validate.py --scratch /tmp/alias-check --run '^(TestNextSupported|TestNextCorpus|TestNextOptions|TestNextThroughput|TestOwnedWitnesses)$'
python3 stage1/cohere/lint/rules/typescript-no-this-alias/validate.py --scratch /tmp/alias-mutant --run '^TestMutants$/^parenthesized_alias_lost$'
```

Redirect wrapper output to logs. Raw test logs live in evidence/*.txt.
The report at ../../claims/wave1-07-third-report.md states counts and limitations.


## Owner harness completion

Default .a discovery is now available through the owner handoff merged at f22d2aa1.
The earlier scratch-overlay instructions are historical. Fresh validation uses
validate-owned.py --scratch <directory> --compiler /tmp/wave07-typescript;
see evidence-completion and ../typescript-no-non-null-assertion/COMPLETION.md.
