# Enum tag kind aliases (#s8gy89r U7, an application of #00rd54t)

tsc's checkExpressionWorker saves `const kind = node.kind;` and switches on
kind, and the checker narrows node by that switch as by `switch (node.kind)`.
The walk's switch-clause step now asks switchField: directField of the switch's
expression, or kindAliasField's answer for a name. kindAliasField returns the
field only when the name is one unannotated const (not let, var or using) whose
initializer fieldRead accepts on the walk's reference, when its statement is an
earlier statement of a list on the switch's road out with only blocks, ifs,
switches, case clauses, labels, try and catch between (no function, class or
loop), and when no name for the reference's binding is written from the alias's
end to the switch's start. A reference that isn't a plain name gets "". If
conditions over the alias are unchanged, and rule 3350 judges every read.

Each producer is a top-level TestProduct_KindAlias<Name>Mutant: t.Parallel(),
one buildcache.GoTestMutated build, no subtests and no other assertions.
TestKindAliasMutantsAreCaught runs the named rows in the mutant binary and
requires each row's direct recognizer witness, so a later lowering stop cannot
hide a fail-open recognizer. The rows also run unmutated in the focused suite
and in the whole internal/lower.

| Mutant | Condition | Catching test | Catch line in enum_tag_kind_alias_test.go |
| --- | --- | --- | --- |
| Initializer: take the initializer's name without fieldRead | 1 | TestKindAliasFallbacksKeepFamilyOne/(other_reference|getter) | 101, `kind-alias fallback <row> was recognized` |
| RunsFirst: skip kindAliasRunsFirst | 2 | TestKindAliasFallbacksKeepFamilyOne/(after|other_function|loop) | 101, `kind-alias fallback <row> was recognized` |
| Rewritten: skip kindAliasRewritten | 3 | TestKindAliasFallbacksKeepFamilyOne/(assigned|destructured|shorthand) | 101, `kind-alias fallback <row> was recognized` |
| Let: accept any block-scoped alias | 1, the let row | TestKindAliasFallbacksKeepFamilyOne/let | 101, `kind-alias fallback let was recognized` |

The catcher requires every listed row's witness, not one failing row.

Rows the checker narrows by their alias (after, other_function, loop, getter,
assigned, destructured, shorthand) are also refused at full Lower with the
family 1 words, as they are on main ac79b0b5. The loop row holds an alias
outside a loop whose body writes the reference after the switch, which the
back edge carries into the next switch; condition 3's source range can't see
it, so condition 2 refuses any loop between.

The checker narrows nothing by a let alias or by an alias of another reference
(getCandidateDiscriminantPropertyAccess asks for a const whose initializer reads
the reference): TestKindAliasCheckerIgnoresLetAndOtherReference pins TS2339 for
both. So no let or other-reference program reaches the walk through its alias.
Those two rows read n.a under their own comparison, which proves it on main and
here alike; they agree with Node on JavaScript and sanitized native, and only
the direct recognizer line can see the Initializer and Let mutants there.

Step zero ran TestKindSetPredicateLieStopsAtCheckedRead with a switch in place
of the predicate, directly and through an alias. Its union has no open tag, so
the walk is never asked, and both forms behave alike on main ac79b0b5 and here.
Native's crash there is #svtsk2x, another owner's; no U7 test pins it.

TestKindAliasKeepsRule3350: a module-level let written by swap(), called
between the alias and the read, is recognized by the alias and then stopped
with 3350's NotYet words, as the direct switch is on main.
