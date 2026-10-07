# transformers/module/module.ts: two remaining declines

All-code pinned latent census **4 -> 2**. The current file was byte-identical to
the previously reviewed Wave C input before these edits. Two U-endpoint reads
become assertions at their original evaluation points. The dynamic-import
callback is reached only for an import call with at least one argument or a
require call with exactly one argument. Its populated AST argument array is
unchanged before the first read. The second read already has predicate narrowing
and gets no added assertion. The JSON first-statement read is immediately guarded
by statements.length and uses a populated JSON NodeArray.

Every remaining finding is in after.json. At line 2262, the SourceFile emission
notification is unconditional, but transformSourceFile skips ordinary scripts
before populating moduleInfoMap. The payload can be absent; asserting it would
be false. The owner currentModuleInfo could be widened, but that requires proofs
for the numerous transformation and substitution reads of that context variable.
That complete protocol proof is unfinished, so this is an honest decline.
At line 2481, arrayFrom(Set<Identifier>) acquires an undefined element through
overload inference. This is a contract diagnostic, not a required indexed read.
An explicit type argument is a plausible type-only repair; it has not been
validated against the native overload resolver and is declined rather than cast.

Stock JavaScript equality and idempotence pass for all 30 owned files. The JSON
endpoint ! -> ?? 0 mutant fails both the emitted-JavaScript and site-contract
checks. Census command: latent-file-census.sh /tmp/emit33-close-meter
/tmp/emit33-close-meter-source /tmp/emit33-close-latent-module. Verification uses
/tmp/emit33-close-source-before-module as the before snapshot. Default oracle
and mechanical API projection results accompany this proof.
