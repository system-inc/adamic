# Direct public-declaration references

Stock TypeScript symbol resolution over all 712 src TypeScript/TSX roots; all 526 explicit accesses and destructuring entries, including homonyms, are retained in the JSON inventory. This table includes writes as well as reads. It is a static inventory, not lane coverage.

| Source | Expression |
|---|---|
| src/compiler/sys.ts:1385:31 | `sys.writeFile` |
| src/compiler/sys.ts:1386:5 | `sys.writeFile` |
| src/compiler/moduleNameResolver.ts:820:9 | `host.directoryExists` |
| src/compiler/moduleNameResolver.ts:824:21 | `host.directoryExists` |
| src/compiler/program.ts:460:14 | `compilerHost.directoryExists` |
| src/compiler/program.ts:478:49 | `system.writeFile` |
| src/compiler/program.ts:4895:37 | `host.compilerHost.directoryExists` |
| src/compiler/program.ts:4908:27 | `host.compilerHost.directoryExists` |
| src/compiler/program.ts:4957:9 | `host.compilerHost.directoryExists` |
| src/compiler/watch.ts:859:56 | `system.writeFile` |
| src/compiler/watchPublic.ts:660:9 | `compilerHost.directoryExists` |
| src/compiler/executeCommandLine.ts:1300:9 | `sys.writeFile` |
| src/services/services.ts:1811:35 | `compilerHost!.directoryExists` |
| src/server/scriptInfo.ts:659:9 | `this.host.writeFile` |
| src/server/project.ts:783:16 | `this.projectService.host.writeFile` |
| src/server/session.ts:2632:111 | `this.host.writeFile` |
| src/harness/harnessIO.ts:126:39 | `ts.sys.writeFile` |
| src/testRunner/unittests/helpers/baseline.ts:137:9 | `sys.writeFile` |
| src/testRunner/unittests/helpers/baseline.ts:294:5 | `sys.writeFile` |
| src/testRunner/unittests/helpers/baseline.ts:460:31 | `sys.writeFile` |
| src/testRunner/unittests/helpers/baseline.ts:461:5 | `sys.writeFile` |
| src/testRunner/unittests/sys/symlinkWatching.ts:76:33 | `sys.writeFile` |
| src/testRunner/unittests/sys/symlinkWatching.ts:336:25 | `sys.writeFile` |
| src/testRunner/unittests/sys/symlinkWatching.ts:340:25 | `sys.writeFile` |
| src/testRunner/unittests/sys/symlinkWatching.ts:344:25 | `sys.writeFile` |
| src/testRunner/unittests/sys/symlinkWatching.ts:575:21 | `sys.writeFile` |
| src/testRunner/unittests/sys/symlinkWatching.ts:576:21 | `sys.writeFile` |
| src/testRunner/unittests/sys/symlinkWatching.ts:583:21 | `sys.writeFile` |
| src/testRunner/unittests/sys/symlinkWatching.ts:590:21 | `sys.writeFile` |
| src/testRunner/unittests/sys/symlinkWatching.ts:693:17 | `ts.sys.writeFile` |
| src/testRunner/unittests/sys/symlinkWatching.ts:706:17 | `ts.sys.writeFile` |
| src/testRunner/unittests/sys/symlinkWatching.ts:719:17 | `ts.sys.writeFile` |
| src/testRunner/unittests/sys/symlinkWatching.ts:733:17 | `ts.sys.writeFile` |
| src/testRunner/unittests/sys/symlinkWatching.ts:763:13 | `ts.sys.writeFile` |
| src/testRunner/unittests/sys/symlinkWatching.ts:764:13 | `ts.sys.writeFile` |
