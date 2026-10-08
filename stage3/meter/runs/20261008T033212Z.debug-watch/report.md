Main 132a0ed5 debug/watch own-file diagnosis.
Compilers: main 132a0ed5b4bec3dc7bc838d9dfec882e5ff9505a; scratch library merge a609ec184bede2216c81d63230863473d75ca1e3 (library 24d980a7ef594cdf0b9c532f680d5b348c3a4b80).
Census: baseline own-file debug 2, watch 2; two-site revert debug 0; replant debug 2.
Library: the four original diagnostics disappear, but debug and watch each gain one TS6307.
Scope: checker diagnosis only; no native execution or full meter run.

Both debug errors are adaptation 40/error-host sites from e17f5248, lines 200 and 201.
Only those two Error receivers were reverted; installed dependencies alone leave debug at 2.
The revert makes debug pass own-file, while its imported program remains checker-rejected.
Library 24d980a7 does not make either requested file pass own-file in the existing single-root census.
Stock TypeScript 6.0.3 checks all 79 project files with zero diagnostics; each single-root check gains TS6307.
The project has composite enabled. The library loader substitutes the requested roots via config.WithFileNames(roots), losing the project file list.

Baseline diagnostics, verbatim:
```
/workspace/stage3-diagnosis-adapted/src/compiler/debug.ts:200:19: error TS2339: Property 'captureStackTrace' does not exist on type 'ErrorConstructor'.
/workspace/stage3-diagnosis-adapted/src/compiler/debug.ts:201:19: error TS2339: Property 'captureStackTrace' does not exist on type 'ErrorConstructor'.
/workspace/stage3-diagnosis-adapted/src/compiler/watch.ts:756:11: error TS2375: Type '{ getSourceFile: (fileName: string, languageVersionOrOptions: CreateSourceFileOptions | ScriptTarget, onError?: ((message: string) => void) | undefined, shouldCreateNewSourceFile?: boolean | undefined) => SourceFile | undefined; ... 17 more ...; jsDocParsingMode: JSDocParsingMode | undefined; }' is not assignable to type 'CompilerHost' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
  Types of property 'getDefaultLibLocation' are incompatible.
    Type '(() => string) | undefined' is not assignable to type '() => string'.
      Type 'undefined' is not assignable to type '() => string'.
/workspace/stage3-diagnosis-adapted/src/compiler/watch.ts:845:5: error TS2375: Type '{ useCaseSensitiveFileNames: () => boolean; getNewLine: () => string; getCurrentDirectory: () => string; getDefaultLibLocation: () => string; getDefaultLibFileName: (options: CompilerOptions) => string; ... 13 more ...; now: ... | undefined; }' is not assignable to type 'ProgramHost<T>' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
  Types of property 'createHash' are incompatible.
    Type '((data: string) => string) | undefined' is not assignable to type '(data: string) => string'.
      Type 'undefined' is not assignable to type '(data: string) => string'.
```

Library own-file diagnostics are retained in full in diagnosis.json.
The original adapted debug/watch source hashes are unchanged after the library check.

Reproduction commands:
```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > setup.log 2>&1
source /workspace/adamic-tools/env.sh
bash stage3/apply.sh /workspace/stage3-diagnosis-adapted > apply.log 2>&1
go build -o /workspace/stage3-diagnosis-main-census ./stage3/census/tool > main-build.log 2>&1
/workspace/stage3-diagnosis-main-census /workspace/stage3-diagnosis-adapted/src/compiler/debug.ts debug-baseline.jsonl > debug-baseline.log 2>&1
/workspace/stage3-diagnosis-main-census /workspace/stage3-diagnosis-adapted/src/compiler/watch.ts watch-baseline.jsonl > watch-baseline.log 2>&1
# Apply two-site-revert.patch to a separate tree and rerun the debug root; restore and rerun for the replant mutant.
npm ci --prefix /workspace/stage3-diagnosis-adapted --ignore-scripts --no-audit --no-fund > project-install.log 2>&1
# Rerun main debug as the dependencies-only control.
git merge --no-edit 24d980a7 > library-merge.log 2>&1
go build -o /workspace/stage3-diagnosis-library-census ./stage3/census/tool > library-build.log 2>&1
# Run both unchanged adapted roots with the new library census.
```

Setup succeeded in 1088.051s, nproc=5, CPU quota=4, Node 24.19.0; timing lines are in setup.txt.
