# Public member references across src

The complete 526-entry JSON inventory records explicit accesses and destructuring, symbol declaration origins, and unrelated homonyms. It resolves all 712 adapted src TypeScript/TSX roots, including harness and test runners. This table includes writes. System.directoryExists is included to price the alternate public declaration interpretation.

| Member | Source | Expression |
|---|---|---|
| System.writeFile | src/compiler/sys.ts:1385:31 | `sys.writeFile` |
| System.writeFile | src/compiler/sys.ts:1386:5 | `sys.writeFile` |
| System.directoryExists | src/compiler/sys.ts:1393:21 | `sys.directoryExists` |
| System.directoryExists | src/compiler/sys.ts:1540:22 | `nodeSystem.directoryExists` |
| ModuleResolutionHost.directoryExists | src/compiler/moduleNameResolver.ts:820:9 | `host.directoryExists` |
| ModuleResolutionHost.directoryExists | src/compiler/moduleNameResolver.ts:824:21 | `host.directoryExists` |
| ModuleResolutionHost.directoryExists | src/compiler/program.ts:460:14 | `compilerHost.directoryExists` |
| System.directoryExists | src/compiler/program.ts:460:46 | `system.directoryExists` |
| System.writeFile | src/compiler/program.ts:478:49 | `system.writeFile` |
| System.directoryExists | src/compiler/program.ts:489:43 | `system.directoryExists` |
| ModuleResolutionHost.directoryExists | src/compiler/program.ts:4895:37 | `host.compilerHost.directoryExists` |
| ModuleResolutionHost.directoryExists | src/compiler/program.ts:4908:27 | `host.compilerHost.directoryExists` |
| ModuleResolutionHost.directoryExists | src/compiler/program.ts:4957:9 | `host.compilerHost.directoryExists` |
| System.directoryExists | src/compiler/watch.ts:852:34 | `system.directoryExists` |
| System.writeFile | src/compiler/watch.ts:859:56 | `system.writeFile` |
| ModuleResolutionHost.directoryExists | src/compiler/watchPublic.ts:660:9 | `compilerHost.directoryExists` |
| System.directoryExists | src/compiler/executeCommandLine.ts:609:61 | `sys.directoryExists` |
| System.writeFile | src/compiler/executeCommandLine.ts:1300:9 | `sys.writeFile` |
| ModuleResolutionHost.directoryExists | src/services/services.ts:1811:35 | `compilerHost!.directoryExists` |
| System.writeFile | src/server/scriptInfo.ts:659:9 | `this.host.writeFile` |
| System.writeFile | src/server/project.ts:783:16 | `this.projectService.host.writeFile` |
| System.writeFile | src/server/session.ts:2632:111 | `this.host.writeFile` |
| System.directoryExists | src/harness/harnessIO.ts:113:23 | `ts.sys.directoryExists` |
| System.writeFile | src/harness/harnessIO.ts:126:39 | `ts.sys.writeFile` |
| System.directoryExists | src/harness/harnessIO.ts:131:34 | `ts.sys.directoryExists` |
| System.writeFile | src/testRunner/unittests/helpers/baseline.ts:137:9 | `sys.writeFile` |
| System.writeFile | src/testRunner/unittests/helpers/baseline.ts:294:5 | `sys.writeFile` |
| System.writeFile | src/testRunner/unittests/helpers/baseline.ts:460:31 | `sys.writeFile` |
| System.writeFile | src/testRunner/unittests/helpers/baseline.ts:461:5 | `sys.writeFile` |
| System.writeFile | src/testRunner/unittests/sys/symlinkWatching.ts:76:33 | `sys.writeFile` |
| System.writeFile | src/testRunner/unittests/sys/symlinkWatching.ts:336:25 | `sys.writeFile` |
| System.writeFile | src/testRunner/unittests/sys/symlinkWatching.ts:340:25 | `sys.writeFile` |
| System.writeFile | src/testRunner/unittests/sys/symlinkWatching.ts:344:25 | `sys.writeFile` |
| System.writeFile | src/testRunner/unittests/sys/symlinkWatching.ts:575:21 | `sys.writeFile` |
| System.writeFile | src/testRunner/unittests/sys/symlinkWatching.ts:576:21 | `sys.writeFile` |
| System.writeFile | src/testRunner/unittests/sys/symlinkWatching.ts:583:21 | `sys.writeFile` |
| System.writeFile | src/testRunner/unittests/sys/symlinkWatching.ts:590:21 | `sys.writeFile` |
| System.writeFile | src/testRunner/unittests/sys/symlinkWatching.ts:693:17 | `ts.sys.writeFile` |
| System.writeFile | src/testRunner/unittests/sys/symlinkWatching.ts:706:17 | `ts.sys.writeFile` |
| System.writeFile | src/testRunner/unittests/sys/symlinkWatching.ts:719:17 | `ts.sys.writeFile` |
| System.writeFile | src/testRunner/unittests/sys/symlinkWatching.ts:733:17 | `ts.sys.writeFile` |
| System.writeFile | src/testRunner/unittests/sys/symlinkWatching.ts:763:13 | `ts.sys.writeFile` |
| System.writeFile | src/testRunner/unittests/sys/symlinkWatching.ts:764:13 | `ts.sys.writeFile` |

## Implicit readers

The computed/rest/reflection audit found these additional reads of ModuleResolutionHost.directoryExists through object spread:

| Source | Expression |
|---|---|
| src/harness/incrementalUtils.ts:525 | `...compilerHost` |
| src/testRunner/unittests/helpers/tscWatch.ts:243 | `...compilerHost` |
| src/testRunner/unittests/reuseProgramStructure.ts:305 | `...host` |

No further checker-resolved computed, rest, in/own-presence, or standard Object/Reflect/JSON reflection access exposed these public declarations. The full named inventory retains overrides, homonyms and unresolved receivers; the compressed lexical ledger retains every literal mention, including type references and test strings. These are static src inventories, not claims about arbitrary reflective JavaScript outside src. Candidate-only readers are shown in the exact local diffs.
