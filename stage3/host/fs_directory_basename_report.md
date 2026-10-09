Built: node:path.basename with omitted, string or undefined suffix, using Node v24.19.0's POSIX UTF-16 scan; declarations remain the shared @types/node 25.3.3.
Commits: separate second piece after e5dba453abd9fd6ffb81a81f133c4f0e383f18cc; pushed SHA is reported with the commit.
Checks: lowering PASS 121.837s, native PASS 363.447s, flow PASS 299.080s, freshness PASS 124.683s; both-backend edge fixture and directory oracle PASS, complete Linux counts regeneration PASS 65.556s.
Mutants: ignoring suffix and returning dirname are both caught only by Node stdout; each exits 0 with no ASan/UBSan findings or leaks.
Uncovered: fixture 15's basename dependency is implemented, but the area SHA with its fs-file driver APIs has not arrived, so fixture 15 is not claimed green. No full repository gate or macOS execution.

The new fixture crosses 30 paths and 20 suffixes, and also checks implicit/explicit undefined and a string-or-undefined variable. It includes empty paths, dot and parent segments without normalization, all-slash paths, repeated and trailing slashes, dotfiles, complete and partial suffix matches, case mismatches, backslashes, NUL bytes, Unicode and lone surrogate/paired-surrogate halves. Return lengths as well as output text are compared. This holds Node's unusual distinction between basename('foo','foo') returning empty and basename('/foo','foo') returning foo, and its all-slash suffix behavior.

The runtime ports the POSIX basename scan from Node's own installed v24.19.0 lib/path.js, recorded in the symlink piece's Node source log. UTF-16 units rather than UTF-8 byte ends allow a suffix to match half a surrogate pair. The Node MIT notice is extended to name this port. node_path.c now defines _DARWIN_C_SOURCE after _POSIX_C_SOURCE. The JavaScript backend's existing generic path dispatch calls Node directly.

Supported signatures, import aliasing and a string-or-undefined function parameter are tested. Spread arguments remain named NotYet; win32/posix module objects remain unbuilt and refused through the shared member registry. No unknown member or overload is silently lowered.

All commands source /workspace/adamic-tools/env.sh, and write directly to [logs/fs_directory_basename](logs/fs_directory_basename/):

```
go test ./internal/oracle -run 'TestNodePathBasenameMutants|TestInputAgreesWithNode/internal/oracle/testdata/node_path_basename[.]a$' -count=1 -v
go test ./internal/lower ./internal/native ./internal/javascript ./internal/flow ./internal/fresh -count=1 -timeout 30m -parallel 3
go test ./internal/oracle -run 'TestNodePathBasenameMutants|TestNodeFSDirectorySymlinkMutants|TestInputAgreesWithNode/internal/oracle/testdata/node_(fs_directory|path_)' -count=1 -timeout 30m -parallel 3 -v
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -parallel 3 -args -update-counts
```

The first oracle command passes in 22.877s, the final directory/path regression oracle in 8.592s. JavaScript has no standalone package tests and is checked by both-backend fixture comparisons. Every allocation in the new counts row is freed. The directory System row changes because the directory input now includes another fixture file. No rebase, force-push or main push.
