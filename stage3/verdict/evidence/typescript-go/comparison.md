# Compiler agreement

| Suite | Left oracle passes | Right oracle passes | Agreement | Disagreement |
|---|---:|---:|---:|---:|
| acceptance | 301/301 | 60/301 | 60 | 241 |
| tiny | 1/1 | 0/1 | 0 | 1 |
| baselines | 13453/13453 | 11212/13453 | 11212 | 2241 |

Agreement uses the verdict diagnostic projection for baselines and raw streams for driver suites. Both compilers are also independently judged against unchanged Node expectations. Two identical wrong outputs do not produce success. Tiny repeats one acceptance project.

| Primary disagreement cause | Runs |
|---|---:|
| unsupported CLI option/value | 1361 |
| exit status | 524 |
| checker diagnostic set | 407 |
| message wording/presentation | 164 |
| diagnostic locations | 27 |
| diagnostic ordering | 0 |
| stderr | 0 |
| timeout | 0 |
| non-diagnostic output | 0 |

Cause classification describes observed outputs. A changed diagnostic code/location/set can reflect checker, library, version, option-default or resolver behavior; it does not alone establish a checker bug. Secondary exit and stderr differences remain recorded. Ordering requires equal multisets of complete diagnostic blocks; wording requires equal header multisets. Unsupported options take precedence because they can prevent checking. Exact comparisons never use these classification relaxations.

## exit status

- acceptance: `061_constEnumErrors` ``; streams: exit
- acceptance: `062_interfaceDeclaration1` ``; streams: exit
- acceptance: `063_didYouMeanSuggestionErrors` ``; streams: exit

## message wording/presentation

- acceptance: `064_recursiveFunctionTypes` ``; streams: stdout, exit
- acceptance: `080_builtinIterator` ``; streams: stdout, exit
- acceptance: `097_mappedTypeGenericWithKnownKeys` ``; streams: stdout, exit

## diagnostic locations

- acceptance: `082_unusedDestructuring` ``; streams: stdout, exit
- acceptance: `178_bitwiseCompoundAssignmentOperators` ``; streams: stdout, exit
- baselines: `tests/cases/compiler/bitwiseCompoundAssignmentOperators.ts` ``; streams: stdout

## checker diagnostic set

- acceptance: `106_interfaceMergeWithNonGenericTypeArguments` ``; streams: stdout, exit
- acceptance: `116_recursivelyExpandingUnionNoStackoverflow` ``; streams: stdout, exit
- acceptance: `177_readonlyTupleAndArrayElaboration` ``; streams: stdout, exit

## unsupported CLI option/value

- baselines: `tests/cases/compiler/SystemModuleForStatementNoInitializer.ts` ``; streams: stdout
- baselines: `tests/cases/compiler/abstractPropertyBasics.ts` `target=es5`; streams: stdout
- baselines: `tests/cases/compiler/accessorWithLineTerminator.ts` `target=es5`; streams: stdout

