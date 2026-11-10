# Full-mode single-file latent census site lists

All three runs use LATENT_FULL=1, the exact LATENT_ONLY_FILE semver path, and LATENT_ASSERT_NO_OUTPUT=1. The meter overlay generator and latent tool are unchanged. All three exit 0, have empty checker diagnostics, and end with the relation trailer. These are measurement findings, not native acceptance.

Every finding occurrence is listed below in its original emission order, including rollback boundaries and imported debug.ts findings. The complete declarations, units, fields and relation trailer remain in each raw JSONL.

## main

29 units, 60 findings.

| # | Phase | Kind | Site | Unit | Reason |
| --- | --- | --- | --- | --- | --- |
| 1 | refusal_scan | Refused | semver.ts:287:34 | semver.ts:276:1 | a cast the runtime can't check |
| 2 | lowering | NotYet | debug.ts:227:5 | semver.ts:46:1 | checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; } |
| 3 | lowering | Boundary | debug.ts:235:9 | semver.ts:46:1 | failed statement; state rolled back; continued at next statement or declaration |
| 4 | lowering | NotYet | debug.ts:252:33 | semver.ts:46:1 | a value of type never |
| 5 | lowering | Boundary | semver.ts:118:17 | semver.ts:46:1 | failed statement; state rolled back; continued at next statement or declaration |
| 6 | lowering | NotYet | semver.ts:124:13 | semver.ts:46:1 | a destructured name that isn't plain |
| 7 | lowering | Boundary | semver.ts:123:9 | semver.ts:46:1 | failed statement; state rolled back; continued at next statement or declaration |
| 8 | lowering | NotYet | semver.ts:130:28 | semver.ts:46:1 | reading major |
| 9 | lowering | Boundary | semver.ts:130:9 | semver.ts:46:1 | failed statement; state rolled back; continued at next statement or declaration |
| 10 | lowering | Refused | semver.ts:47:41 | semver.ts:46:1 | a constructor object escaping before static fields are initialized |
| 11 | lowering | Boundary | semver.ts:46:1 | semver.ts:46:1 | failed statement; state rolled back; continued at next statement or declaration |
| 12 | lowering | NotYet | debug.ts:227:5 | semver.ts:207:1 | checked predicate overload target Comparator[][] |
| 13 | lowering | Boundary | debug.ts:235:9 | semver.ts:207:1 | failed statement; state rolled back; continued at next statement or declaration |
| 14 | lowering | NotYet | debug.ts:227:5 | semver.ts:207:1 | checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; } |
| 15 | lowering | Boundary | debug.ts:235:9 | semver.ts:207:1 | failed statement; state rolled back; continued at next statement or declaration |
| 16 | lowering | NotYet | debug.ts:252:33 | semver.ts:207:1 | a value of type never |
| 17 | lowering | Boundary | semver.ts:118:17 | semver.ts:207:1 | failed statement; state rolled back; continued at next statement or declaration |
| 18 | lowering | NotYet | semver.ts:124:13 | semver.ts:207:1 | a destructured name that isn't plain |
| 19 | lowering | Boundary | semver.ts:123:9 | semver.ts:207:1 | failed statement; state rolled back; continued at next statement or declaration |
| 20 | lowering | NotYet | semver.ts:130:28 | semver.ts:207:1 | reading major |
| 21 | lowering | Boundary | semver.ts:130:9 | semver.ts:207:1 | failed statement; state rolled back; continued at next statement or declaration |
| 22 | lowering | Refused | semver.ts:287:34 | semver.ts:276:1 | a cast the runtime can't check |
| 23 | lowering | Boundary | semver.ts:287:13 | semver.ts:276:1 | failed statement; state rolled back; continued at next statement or declaration |
| 24 | lowering | NotYet | debug.ts:227:5 | semver.ts:297:1 | checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; } |
| 25 | lowering | Boundary | debug.ts:235:9 | semver.ts:297:1 | failed statement; state rolled back; continued at next statement or declaration |
| 26 | lowering | NotYet | debug.ts:252:33 | semver.ts:297:1 | a value of type never |
| 27 | lowering | Boundary | semver.ts:118:17 | semver.ts:297:1 | failed statement; state rolled back; continued at next statement or declaration |
| 28 | lowering | NotYet | semver.ts:124:13 | semver.ts:297:1 | a destructured name that isn't plain |
| 29 | lowering | Boundary | semver.ts:123:9 | semver.ts:297:1 | failed statement; state rolled back; continued at next statement or declaration |
| 30 | lowering | NotYet | semver.ts:130:28 | semver.ts:297:1 | reading major |
| 31 | lowering | Boundary | semver.ts:130:9 | semver.ts:297:1 | failed statement; state rolled back; continued at next statement or declaration |
| 32 | lowering | NotYet | debug.ts:227:5 | semver.ts:314:1 | checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; } |
| 33 | lowering | Boundary | debug.ts:235:9 | semver.ts:314:1 | failed statement; state rolled back; continued at next statement or declaration |
| 34 | lowering | NotYet | debug.ts:252:33 | semver.ts:314:1 | a value of type never |
| 35 | lowering | Boundary | semver.ts:118:17 | semver.ts:314:1 | failed statement; state rolled back; continued at next statement or declaration |
| 36 | lowering | NotYet | semver.ts:124:13 | semver.ts:314:1 | a destructured name that isn't plain |
| 37 | lowering | Boundary | semver.ts:123:9 | semver.ts:314:1 | failed statement; state rolled back; continued at next statement or declaration |
| 38 | lowering | NotYet | semver.ts:130:28 | semver.ts:314:1 | reading major |
| 39 | lowering | Boundary | semver.ts:130:9 | semver.ts:314:1 | failed statement; state rolled back; continued at next statement or declaration |
| 40 | lowering | NotYet | debug.ts:227:5 | semver.ts:336:1 | checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; } |
| 41 | lowering | Boundary | debug.ts:235:9 | semver.ts:336:1 | failed statement; state rolled back; continued at next statement or declaration |
| 42 | lowering | NotYet | debug.ts:252:33 | semver.ts:336:1 | a value of type never |
| 43 | lowering | Boundary | semver.ts:118:17 | semver.ts:336:1 | failed statement; state rolled back; continued at next statement or declaration |
| 44 | lowering | NotYet | semver.ts:124:13 | semver.ts:336:1 | a destructured name that isn't plain |
| 45 | lowering | Boundary | semver.ts:123:9 | semver.ts:336:1 | failed statement; state rolled back; continued at next statement or declaration |
| 46 | lowering | NotYet | semver.ts:130:28 | semver.ts:336:1 | reading major |
| 47 | lowering | Boundary | semver.ts:130:9 | semver.ts:336:1 | failed statement; state rolled back; continued at next statement or declaration |
| 48 | lowering | NotYet | semver.ts:380:13 | semver.ts:336:1 | a case whose type differs from the switch's |
| 49 | lowering | Boundary | semver.ts:342:9 | semver.ts:336:1 | failed statement; state rolled back; continued at next statement or declaration |
| 50 | lowering | NotYet | debug.ts:227:5 | semver.ts:425:1 | checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; } |
| 51 | lowering | Boundary | debug.ts:235:9 | semver.ts:425:1 | failed statement; state rolled back; continued at next statement or declaration |
| 52 | lowering | NotYet | debug.ts:252:33 | semver.ts:425:1 | a value of type never |
| 53 | lowering | Boundary | semver.ts:118:17 | semver.ts:425:1 | failed statement; state rolled back; continued at next statement or declaration |
| 54 | lowering | NotYet | semver.ts:124:13 | semver.ts:425:1 | a destructured name that isn't plain |
| 55 | lowering | Boundary | semver.ts:123:9 | semver.ts:425:1 | failed statement; state rolled back; continued at next statement or declaration |
| 56 | lowering | NotYet | semver.ts:130:28 | semver.ts:425:1 | reading major |
| 57 | lowering | Boundary | semver.ts:130:9 | semver.ts:425:1 | failed statement; state rolled back; continued at next statement or declaration |
| 58 | lowering | Boundary | semver.ts:439:13 | semver.ts:425:1 | failed statement; state rolled back; continued at next statement or declaration |
| 59 | lowering | NotYet | semver.ts:452:37 | semver.ts:451:1 | a template interpolating an object, an array, a map, a function or undefined |
| 60 | lowering | Boundary | semver.ts:452:5 | semver.ts:451:1 | failed statement; state rolled back; continued at next statement or declaration |

## amended

28 units, 57 findings.

| # | Phase | Kind | Site | Unit | Reason |
| --- | --- | --- | --- | --- | --- |
| 1 | lowering | NotYet | debug.ts:227:5 | semver.ts:46:1 | checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; } |
| 2 | lowering | Boundary | debug.ts:235:9 | semver.ts:46:1 | failed statement; state rolled back; continued at next statement or declaration |
| 3 | lowering | NotYet | debug.ts:252:33 | semver.ts:46:1 | a value of type never |
| 4 | lowering | Boundary | semver.ts:118:17 | semver.ts:46:1 | failed statement; state rolled back; continued at next statement or declaration |
| 5 | lowering | NotYet | semver.ts:124:13 | semver.ts:46:1 | a destructured name that isn't plain |
| 6 | lowering | Boundary | semver.ts:123:9 | semver.ts:46:1 | failed statement; state rolled back; continued at next statement or declaration |
| 7 | lowering | NotYet | semver.ts:130:28 | semver.ts:46:1 | reading major |
| 8 | lowering | Boundary | semver.ts:130:9 | semver.ts:46:1 | failed statement; state rolled back; continued at next statement or declaration |
| 9 | lowering | Refused | semver.ts:47:41 | semver.ts:46:1 | a constructor object escaping before static fields are initialized |
| 10 | lowering | Boundary | semver.ts:46:1 | semver.ts:46:1 | failed statement; state rolled back; continued at next statement or declaration |
| 11 | lowering | NotYet | debug.ts:227:5 | semver.ts:207:1 | checked predicate overload target Comparator[][] |
| 12 | lowering | Boundary | debug.ts:235:9 | semver.ts:207:1 | failed statement; state rolled back; continued at next statement or declaration |
| 13 | lowering | NotYet | debug.ts:227:5 | semver.ts:207:1 | checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; } |
| 14 | lowering | Boundary | debug.ts:235:9 | semver.ts:207:1 | failed statement; state rolled back; continued at next statement or declaration |
| 15 | lowering | NotYet | debug.ts:252:33 | semver.ts:207:1 | a value of type never |
| 16 | lowering | Boundary | semver.ts:118:17 | semver.ts:207:1 | failed statement; state rolled back; continued at next statement or declaration |
| 17 | lowering | NotYet | semver.ts:124:13 | semver.ts:207:1 | a destructured name that isn't plain |
| 18 | lowering | Boundary | semver.ts:123:9 | semver.ts:207:1 | failed statement; state rolled back; continued at next statement or declaration |
| 19 | lowering | NotYet | semver.ts:130:28 | semver.ts:207:1 | reading major |
| 20 | lowering | Boundary | semver.ts:130:9 | semver.ts:207:1 | failed statement; state rolled back; continued at next statement or declaration |
| 21 | lowering | NotYet | debug.ts:227:5 | semver.ts:297:1 | checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; } |
| 22 | lowering | Boundary | debug.ts:235:9 | semver.ts:297:1 | failed statement; state rolled back; continued at next statement or declaration |
| 23 | lowering | NotYet | debug.ts:252:33 | semver.ts:297:1 | a value of type never |
| 24 | lowering | Boundary | semver.ts:118:17 | semver.ts:297:1 | failed statement; state rolled back; continued at next statement or declaration |
| 25 | lowering | NotYet | semver.ts:124:13 | semver.ts:297:1 | a destructured name that isn't plain |
| 26 | lowering | Boundary | semver.ts:123:9 | semver.ts:297:1 | failed statement; state rolled back; continued at next statement or declaration |
| 27 | lowering | NotYet | semver.ts:130:28 | semver.ts:297:1 | reading major |
| 28 | lowering | Boundary | semver.ts:130:9 | semver.ts:297:1 | failed statement; state rolled back; continued at next statement or declaration |
| 29 | lowering | NotYet | debug.ts:227:5 | semver.ts:314:1 | checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; } |
| 30 | lowering | Boundary | debug.ts:235:9 | semver.ts:314:1 | failed statement; state rolled back; continued at next statement or declaration |
| 31 | lowering | NotYet | debug.ts:252:33 | semver.ts:314:1 | a value of type never |
| 32 | lowering | Boundary | semver.ts:118:17 | semver.ts:314:1 | failed statement; state rolled back; continued at next statement or declaration |
| 33 | lowering | NotYet | semver.ts:124:13 | semver.ts:314:1 | a destructured name that isn't plain |
| 34 | lowering | Boundary | semver.ts:123:9 | semver.ts:314:1 | failed statement; state rolled back; continued at next statement or declaration |
| 35 | lowering | NotYet | semver.ts:130:28 | semver.ts:314:1 | reading major |
| 36 | lowering | Boundary | semver.ts:130:9 | semver.ts:314:1 | failed statement; state rolled back; continued at next statement or declaration |
| 37 | lowering | NotYet | debug.ts:227:5 | semver.ts:336:1 | checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; } |
| 38 | lowering | Boundary | debug.ts:235:9 | semver.ts:336:1 | failed statement; state rolled back; continued at next statement or declaration |
| 39 | lowering | NotYet | debug.ts:252:33 | semver.ts:336:1 | a value of type never |
| 40 | lowering | Boundary | semver.ts:118:17 | semver.ts:336:1 | failed statement; state rolled back; continued at next statement or declaration |
| 41 | lowering | NotYet | semver.ts:124:13 | semver.ts:336:1 | a destructured name that isn't plain |
| 42 | lowering | Boundary | semver.ts:123:9 | semver.ts:336:1 | failed statement; state rolled back; continued at next statement or declaration |
| 43 | lowering | NotYet | semver.ts:130:28 | semver.ts:336:1 | reading major |
| 44 | lowering | Boundary | semver.ts:130:9 | semver.ts:336:1 | failed statement; state rolled back; continued at next statement or declaration |
| 45 | lowering | NotYet | semver.ts:380:13 | semver.ts:336:1 | a case whose type differs from the switch's |
| 46 | lowering | Boundary | semver.ts:342:9 | semver.ts:336:1 | failed statement; state rolled back; continued at next statement or declaration |
| 47 | lowering | NotYet | debug.ts:227:5 | semver.ts:425:1 | checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; } |
| 48 | lowering | Boundary | debug.ts:235:9 | semver.ts:425:1 | failed statement; state rolled back; continued at next statement or declaration |
| 49 | lowering | NotYet | debug.ts:252:33 | semver.ts:425:1 | a value of type never |
| 50 | lowering | Boundary | semver.ts:118:17 | semver.ts:425:1 | failed statement; state rolled back; continued at next statement or declaration |
| 51 | lowering | NotYet | semver.ts:124:13 | semver.ts:425:1 | a destructured name that isn't plain |
| 52 | lowering | Boundary | semver.ts:123:9 | semver.ts:425:1 | failed statement; state rolled back; continued at next statement or declaration |
| 53 | lowering | NotYet | semver.ts:130:28 | semver.ts:425:1 | reading major |
| 54 | lowering | Boundary | semver.ts:130:9 | semver.ts:425:1 | failed statement; state rolled back; continued at next statement or declaration |
| 55 | lowering | Boundary | semver.ts:439:13 | semver.ts:425:1 | failed statement; state rolled back; continued at next statement or declaration |
| 56 | lowering | NotYet | semver.ts:452:37 | semver.ts:451:1 | a template interpolating an object, an array, a map, a function or undefined |
| 57 | lowering | Boundary | semver.ts:452:5 | semver.ts:451:1 | failed statement; state rolled back; continued at next statement or declaration |

## mutant

29 units, 60 findings.

| # | Phase | Kind | Site | Unit | Reason |
| --- | --- | --- | --- | --- | --- |
| 1 | refusal_scan | Refused | semver.ts:287:34 | semver.ts:276:1 | a cast the runtime can't check |
| 2 | lowering | NotYet | debug.ts:227:5 | semver.ts:46:1 | checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; } |
| 3 | lowering | Boundary | debug.ts:235:9 | semver.ts:46:1 | failed statement; state rolled back; continued at next statement or declaration |
| 4 | lowering | NotYet | debug.ts:252:33 | semver.ts:46:1 | a value of type never |
| 5 | lowering | Boundary | semver.ts:118:17 | semver.ts:46:1 | failed statement; state rolled back; continued at next statement or declaration |
| 6 | lowering | NotYet | semver.ts:124:13 | semver.ts:46:1 | a destructured name that isn't plain |
| 7 | lowering | Boundary | semver.ts:123:9 | semver.ts:46:1 | failed statement; state rolled back; continued at next statement or declaration |
| 8 | lowering | NotYet | semver.ts:130:28 | semver.ts:46:1 | reading major |
| 9 | lowering | Boundary | semver.ts:130:9 | semver.ts:46:1 | failed statement; state rolled back; continued at next statement or declaration |
| 10 | lowering | Refused | semver.ts:47:41 | semver.ts:46:1 | a constructor object escaping before static fields are initialized |
| 11 | lowering | Boundary | semver.ts:46:1 | semver.ts:46:1 | failed statement; state rolled back; continued at next statement or declaration |
| 12 | lowering | NotYet | debug.ts:227:5 | semver.ts:207:1 | checked predicate overload target Comparator[][] |
| 13 | lowering | Boundary | debug.ts:235:9 | semver.ts:207:1 | failed statement; state rolled back; continued at next statement or declaration |
| 14 | lowering | NotYet | debug.ts:227:5 | semver.ts:207:1 | checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; } |
| 15 | lowering | Boundary | debug.ts:235:9 | semver.ts:207:1 | failed statement; state rolled back; continued at next statement or declaration |
| 16 | lowering | NotYet | debug.ts:252:33 | semver.ts:207:1 | a value of type never |
| 17 | lowering | Boundary | semver.ts:118:17 | semver.ts:207:1 | failed statement; state rolled back; continued at next statement or declaration |
| 18 | lowering | NotYet | semver.ts:124:13 | semver.ts:207:1 | a destructured name that isn't plain |
| 19 | lowering | Boundary | semver.ts:123:9 | semver.ts:207:1 | failed statement; state rolled back; continued at next statement or declaration |
| 20 | lowering | NotYet | semver.ts:130:28 | semver.ts:207:1 | reading major |
| 21 | lowering | Boundary | semver.ts:130:9 | semver.ts:207:1 | failed statement; state rolled back; continued at next statement or declaration |
| 22 | lowering | Refused | semver.ts:287:34 | semver.ts:276:1 | a cast the runtime can't check |
| 23 | lowering | Boundary | semver.ts:287:13 | semver.ts:276:1 | failed statement; state rolled back; continued at next statement or declaration |
| 24 | lowering | NotYet | debug.ts:227:5 | semver.ts:297:1 | checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; } |
| 25 | lowering | Boundary | debug.ts:235:9 | semver.ts:297:1 | failed statement; state rolled back; continued at next statement or declaration |
| 26 | lowering | NotYet | debug.ts:252:33 | semver.ts:297:1 | a value of type never |
| 27 | lowering | Boundary | semver.ts:118:17 | semver.ts:297:1 | failed statement; state rolled back; continued at next statement or declaration |
| 28 | lowering | NotYet | semver.ts:124:13 | semver.ts:297:1 | a destructured name that isn't plain |
| 29 | lowering | Boundary | semver.ts:123:9 | semver.ts:297:1 | failed statement; state rolled back; continued at next statement or declaration |
| 30 | lowering | NotYet | semver.ts:130:28 | semver.ts:297:1 | reading major |
| 31 | lowering | Boundary | semver.ts:130:9 | semver.ts:297:1 | failed statement; state rolled back; continued at next statement or declaration |
| 32 | lowering | NotYet | debug.ts:227:5 | semver.ts:314:1 | checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; } |
| 33 | lowering | Boundary | debug.ts:235:9 | semver.ts:314:1 | failed statement; state rolled back; continued at next statement or declaration |
| 34 | lowering | NotYet | debug.ts:252:33 | semver.ts:314:1 | a value of type never |
| 35 | lowering | Boundary | semver.ts:118:17 | semver.ts:314:1 | failed statement; state rolled back; continued at next statement or declaration |
| 36 | lowering | NotYet | semver.ts:124:13 | semver.ts:314:1 | a destructured name that isn't plain |
| 37 | lowering | Boundary | semver.ts:123:9 | semver.ts:314:1 | failed statement; state rolled back; continued at next statement or declaration |
| 38 | lowering | NotYet | semver.ts:130:28 | semver.ts:314:1 | reading major |
| 39 | lowering | Boundary | semver.ts:130:9 | semver.ts:314:1 | failed statement; state rolled back; continued at next statement or declaration |
| 40 | lowering | NotYet | debug.ts:227:5 | semver.ts:336:1 | checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; } |
| 41 | lowering | Boundary | debug.ts:235:9 | semver.ts:336:1 | failed statement; state rolled back; continued at next statement or declaration |
| 42 | lowering | NotYet | debug.ts:252:33 | semver.ts:336:1 | a value of type never |
| 43 | lowering | Boundary | semver.ts:118:17 | semver.ts:336:1 | failed statement; state rolled back; continued at next statement or declaration |
| 44 | lowering | NotYet | semver.ts:124:13 | semver.ts:336:1 | a destructured name that isn't plain |
| 45 | lowering | Boundary | semver.ts:123:9 | semver.ts:336:1 | failed statement; state rolled back; continued at next statement or declaration |
| 46 | lowering | NotYet | semver.ts:130:28 | semver.ts:336:1 | reading major |
| 47 | lowering | Boundary | semver.ts:130:9 | semver.ts:336:1 | failed statement; state rolled back; continued at next statement or declaration |
| 48 | lowering | NotYet | semver.ts:380:13 | semver.ts:336:1 | a case whose type differs from the switch's |
| 49 | lowering | Boundary | semver.ts:342:9 | semver.ts:336:1 | failed statement; state rolled back; continued at next statement or declaration |
| 50 | lowering | NotYet | debug.ts:227:5 | semver.ts:425:1 | checked predicate overload target { major: number; minor: number; patch: number; prerelease: string; build: string; } |
| 51 | lowering | Boundary | debug.ts:235:9 | semver.ts:425:1 | failed statement; state rolled back; continued at next statement or declaration |
| 52 | lowering | NotYet | debug.ts:252:33 | semver.ts:425:1 | a value of type never |
| 53 | lowering | Boundary | semver.ts:118:17 | semver.ts:425:1 | failed statement; state rolled back; continued at next statement or declaration |
| 54 | lowering | NotYet | semver.ts:124:13 | semver.ts:425:1 | a destructured name that isn't plain |
| 55 | lowering | Boundary | semver.ts:123:9 | semver.ts:425:1 | failed statement; state rolled back; continued at next statement or declaration |
| 56 | lowering | NotYet | semver.ts:130:28 | semver.ts:425:1 | reading major |
| 57 | lowering | Boundary | semver.ts:130:9 | semver.ts:425:1 | failed statement; state rolled back; continued at next statement or declaration |
| 58 | lowering | Boundary | semver.ts:439:13 | semver.ts:425:1 | failed statement; state rolled back; continued at next statement or declaration |
| 59 | lowering | NotYet | semver.ts:452:37 | semver.ts:451:1 | a template interpolating an object, an array, a map, a function or undefined |
| 60 | lowering | Boundary | semver.ts:452:5 | semver.ts:451:1 | failed statement; state rolled back; continued at next statement or declaration |
