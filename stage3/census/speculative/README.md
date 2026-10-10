Recursive planner built; watch matches, createNodeBuilder fails the union admission check.
50 KiB threshold: 882 planned pieces across 14 files; checker has 64, createNodeBuilder eight.
Largest planned piece: 49,302 bytes, 189.813s, peak RSS 1036812KiB.
Whole createNodeBuilder: 971.045s, peak RSS 4182624KiB; all 30,499 nodes visited.
Dropped, duplicated, moved and shifted-node mutants caught; full coverage remains 65/79 files.


| Depth | NotYet | Refused |
|---|---:|---:|
| 0 | 9,232 | 3,073 |
| 1 | 2,635 | 1,000 |
| 2 | 1,097 | 351 |
| 3 | 319 | 174 |
| 4+ | 195 | 87 |
| Total | 13,478 | 4,685 |

| Kind | Root kind (exact reason) | 0 | 1 | 2 | 3 | 4+ | Total |
|---|---|---:|---:|---:|---:|---:|---:|
| NotYet | reading SyntaxKind | 2,038 | 196 | 90 | 1 | 0 | 2,325 |
| NotYet | a generic function as a value | 1,506 | 231 | 45 | 10 | 3 | 1,795 |
| NotYet | an overloaded function as a value | 802 | 105 | 44 | 5 | 3 | 959 |
| Refused | an object refinement using an open numeric enum as a literal tag | 684 | 148 | 13 | 5 | 2 | 852 |
| NotYet | reading CharacterCodes | 558 | 17 | 2 | 0 | 0 | 577 |
| NotYet | a union of differently held members variable a function value captures | 128 | 141 | 81 | 64 | 31 | 445 |
| NotYet | an assignment value to a member | 195 | 122 | 53 | 14 | 2 | 386 |
| NotYet | a PostfixUnaryExpression | 301 | 42 | 8 | 0 | 0 | 351 |
| Refused | a cast the runtime can't check | 131 | 107 | 41 | 24 | 31 | 334 |
| Refused | inherited library member push read as an own field | 213 | 75 | 22 | 15 | 3 | 328 |
| Refused | a type predicate whose return is not proven (return expression is not a trusted check on node) | 325 | 0 | 0 | 0 | 0 | 325 |
| Refused | an unproven predicate argument for parameter test (argument "isExpression") | 207 | 79 | 2 | 0 | 0 | 288 |
| NotYet | a value of type any | 126 | 75 | 56 | 13 | 15 | 285 |
| NotYet | a value of type T | 165 | 98 | 15 | 3 | 0 | 281 |
| NotYet | a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions | 107 | 109 | 43 | 11 | 9 | 279 |
| NotYet | a namespace member without a lowered binding | 115 | 149 | 6 | 2 | 0 | 272 |
| NotYet | a value of type Path | 135 | 61 | 23 | 10 | 3 | 232 |
| NotYet | a union or optional field read in a program with record storage | 144 | 48 | 9 | 1 | 1 | 203 |
| NotYet | this outside a method | 5 | 30 | 120 | 8 | 2 | 165 |
| NotYet | an array of T | 32 | 59 | 35 | 13 | 7 | 146 |

Examined **3,760,099 / 10,009,820 TypeScript source bytes (37.564102%)**, in 65 files. Including non-TypeScript inputs, that is **3,760,099 / 10,615,807 bytes (35.419813%)**. Every AST child in every completed TypeScript file was visited, including checker-rejected bodies. Comments and whitespace are included in those parsed source bytes. Structural and type syntax is examined by traversal and the refusal scanner; it is not passed to a value lowerer.

Unexamined inputs:

- `binder.ts`: 195,107 bytes; per-file measurement incomplete; see stream inventory.
- `builder.ts`: 117,298 bytes; per-file measurement incomplete; see stream inventory.
- `checker.ts`: 3,154,156 bytes; per-file measurement incomplete; see stream inventory.
- `commandLineParser.ts`: 183,733 bytes; per-file measurement incomplete; see stream inventory.
- `diagnosticMessages.generated.json`: 306,771 bytes; non-TypeScript input; no lowering sites.
- `diagnosticMessages.json`: 299,061 bytes; non-TypeScript input; no lowering sites.
- `emitter.ts`: 274,627 bytes; per-file measurement incomplete; see stream inventory.
- `executeCommandLine.ts`: 54,177 bytes; per-file measurement incomplete; see stream inventory.
- `factory/nodeFactory.ts`: 336,646 bytes; per-file measurement incomplete; see stream inventory.
- `factory/utilities.ts`: 76,615 bytes; per-file measurement incomplete; see stream inventory.
- `moduleNameResolver.ts`: 183,618 bytes; per-file measurement incomplete; see stream inventory.
- `parser.ts`: 541,260 bytes; per-file measurement incomplete; see stream inventory.
- `program.ts`: 271,300 bytes; per-file measurement incomplete; see stream inventory.
- `transformers/es2015.ts`: 231,789 bytes; per-file measurement incomplete; see stream inventory.
- `tsbuildPublic.ts`: 114,655 bytes; per-file measurement incomplete; see stream inventory.
- `tsconfig.json`: 155 bytes; non-TypeScript input; no lowering sites.
- `utilities.ts`: 514,740 bytes; per-file measurement incomplete; see stream inventory.

Stock TypeScript independently matched the ancestry depth of 20,558 findings at their actual source AST identities, including dependency attempts; 0 site spans had no exact stock AST match; 1 finding had no compiler-source AST identity. This is an error record from a refusal-scan panic on a computed property name, excluded from the NotYet/Refused table; all classified source sites have exact stock matches. The small control proves exact depth 1 and 2 independently of the corpus ranking.

Other raw attempt records, excluded from the NotYet/Refused site tables: {"error": 30}. See method.md and [raw error inventory](evidence/continuation/ERRORS.json) for their interpretation.

All records use the same whole-compiler project, pinned compiler and adapted source hashes as 76487f84. The cumulative table is recounted from the union of completed records; historical and new tables are not summed. Dependency observations may contribute sites without claiming coverage of their entire source file.

| Coverage | Files | TypeScript bytes | Share |
|---|---:|---:|---:|
| Before | 29 / 79 | 1,323,621 / 10,009,820 | 13.223225% |
| After | 65 / 79 | 3,760,099 / 10,009,820 | 37.564102% |

Files still timed out after a full 600-second budget:

- `binder.ts`
- `builder.ts`
- `checker.ts`
- `commandLineParser.ts`
- `emitter.ts`
- `executeCommandLine.ts`
- `factory/nodeFactory.ts`
- `moduleNameResolver.ts`
- `parser.ts`
- `program.ts`
- `transformers/es2015.ts`
- `tsbuildPublic.ts`
- `utilities.ts`

Other incomplete measurements:

| File | Wall seconds | Peak RSS KiB | Exit | Diagnostic |
|---|---:|---:|---:|---|
| factory/utilities.ts | 82.782 | 1,551,612 | 2 | Go stack overflow with repeated lower.inferTypes frames |

The exclusive four-CPU checker probe timed out after **1800.233s**, peak RSS **2,985,980 KiB**. Its last persisted checkpoint at 1792.399s was in `walk` at `/tmp/b17vtn1/original-adapted/src/compiler/checker.ts:25576:17`: **39/51 top-level statements completed**, 14,807 unique selected-source walker nodes, 83,862 lowerer nodes, and 298,509/298,510 nodes in the scanner/walker/lowerer union. The scanner had already covered almost all AST nodes; this union does not mean the value walk completed. No partial checker observations enter the table.

Four retry workers each use one CPU and a 600-second wall limit, 3GiB RSS watchdog and 6GiB address-space cap. The first checker probe overlaps them; the final four-CPU probe runs after the batch finishes. Two initial SIGKILLs coincided with tmpfs/cgroup memory pressure and recorded OOM kills; inactive scratch moved to disk and a 1GiB Go heap target allowed fresh parser/utilities and checker attempts. All attempts are archived separately. The 130-minute worst-case batch ceiling exceeded the two-hour unit budget; checkpoint 155445a9 was pushed when the measured forecast crossed that boundary. Measurements stop by 23:58:15 UTC and publication by 00:05:15 UTC.

Every cumulative per-file measurement over 120 seconds follows; the separate checker probe is listed separately. RSS is the sampled process-group maximum or child rusage maximum, whichever is larger.

| File | Wall seconds | Peak RSS KiB | Exit |
|---|---:|---:|---:|
| binder.ts | 600.184 | 1,035,304 | 124 |
| builder.ts | 600.193 | 1,035,400 | 124 |
| checker.ts | 600.171 | 2,086,644 | 124 |
| commandLineParser.ts | 600.097 | 1,034,868 | 124 |
| debug.ts | 158.078 | 1,037,380 | 0 |
| emitter.ts | 600.164 | 1,456,356 | 124 |
| executeCommandLine.ts | 600.138 | 1,034,904 | 124 |
| expressionToTypeNode.ts | 375.290 | 1,036,024 | 0 |
| factory/emitHelpers.ts | 127.879 | 1,036,608 | 0 |
| factory/nodeFactory.ts | 600.154 | 2,125,656 | 124 |
| moduleNameResolver.ts | 600.085 | 1,036,120 | 124 |
| moduleSpecifiers.ts | 262.144 | 1,034,960 | 0 |
| parser.ts | 600.211 | 1,036,496 | 124 |
| program.ts | 600.075 | 1,036,000 | 124 |
| resolutionCache.ts | 324.269 | 1,034,592 | 0 |
| sourcemap.ts | 264.433 | 1,035,068 | 0 |
| sys.ts | 169.553 | 1,035,540 | 0 |
| transformers/classFields.ts | 474.553 | 1,036,116 | 0 |
| transformers/declarations.ts | 382.121 | 1,036,376 | 0 |
| transformers/destructuring.ts | 142.457 | 1,035,600 | 0 |
| transformers/es2015.ts | 600.212 | 1,043,996 | 124 |
| transformers/es2018.ts | 126.381 | 1,035,348 | 0 |
| transformers/esDecorators.ts | 299.661 | 1,038,484 | 0 |
| transformers/generators.ts | 312.161 | 1,035,724 | 0 |
| transformers/jsx.ts | 120.456 | 1,035,444 | 0 |
| transformers/module/esnextAnd2015.ts | 181.253 | 1,034,612 | 0 |
| transformers/module/module.ts | 329.642 | 1,036,068 | 0 |
| transformers/module/system.ts | 177.401 | 1,035,736 | 0 |
| transformers/ts.ts | 486.720 | 1,036,176 | 0 |
| transformers/utilities.ts | 161.942 | 1,035,860 | 0 |
| tsbuildPublic.ts | 600.122 | 1,035,120 | 124 |
| utilities.ts | 600.157 | 1,035,780 | 124 |
| visitorPublic.ts | 254.085 | 1,036,804 | 0 |
| watch.ts | 503.188 | 1,036,784 | 0 |
| watchPublic.ts | 327.426 | 1,035,412 | 0 |
| checker.ts (separate 1,800s probe) | 1800.233 | 2,985,980 | 124 |

Focused checks: `audit_retry.py`, `audit_progress.py`, `audit_stream.py`, `isolation.py`, `finalize_stream.py` and `verify.py` pass. Mutants caught: serial workers, future deadline, corrupted checksum, dropped completed record, shifted progress nodes/bytes, shifted depth, dropped stream file, incorrect depth/top20/coverage/byte totals, and altered overlay-off C/JS bytes. The prior header controls, mapper Node/native witness and f6bb0b41 directory comparison remain historical evidence; see [prior report](evidence/continuation/BEFORE-README.md) and [validation](validation.md).

See [continuation method](continuation.md), [coverage and every runtime](evidence/continuation/COVERAGE.json), and [archived logs and records](evidence/continuation). SHA256.json inventories every archived file. Setup completed in 15.926s; nproc=5 with a four-CPU quota.

Finalization took 61.210s with peak RSS 2540296KiB. Its whole-project boundary union recalibrated 280 raw depth tags before the independent recount. Published totals contain 18163 unique classified sites. All 50 retries received their intended budgets; no file was shortened by the two-hour deadline.

Task #b2wbbha: [piece method and reproduction](pieces.md), [union, measurements and scaling report](evidence/pieces/README.md). The six checker samples are excluded from this table.

Recursive continuation: [result and failed admission](evidence/recursive-pieces/README.md), [every createNodeBuilder difference](evidence/recursive-pieces/DIFFERENCES.md). No recursive observations enter the depth table.
