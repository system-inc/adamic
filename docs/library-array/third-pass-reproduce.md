# Array third pass reproduction

Compiler base: codex/library-array-2 bcf803d. Claim commits 556c33a and de2a5ad
precede implementation. The separate runner is unchanged af12899. Test262 is
c8c798898646638cd0c24879f8e0374e847e7d74; stock tsc is 6.0.3.

```sh
bash cloud/setup.sh > /tmp/array3-setup.log 2>&1
source /workspace/adamic-tools/env.sh
nproc
# Fetch the two non-main refs explicitly when the clone refspec only tracks main.
git fetch origin codex/library-array-2:refs/remotes/origin/codex/library-array-2 codex/test262-ts-validity:refs/remotes/origin/codex/test262-ts-validity
git worktree add --detach /workspace/array2-validity origin/codex/test262-ts-validity
# Initialize the runner checkout's cohere submodule before building.
go build -C /workspace/array2-validity -buildvcs=false -o /workspace/array3-runner ./cmd/adamic-test262
PATH=/workspace/array2-tsc/node_modules/.bin:$PATH /workspace/array3-runner -adapt -json -root /workspace/adamic -test262 /workspace/test262 -work /tmp/array3-final built-ins/Array > /tmp/array3-final.json 2> /tmp/array3-final.log

go test ./internal/lower ./internal/native ./internal/javascript ./internal/ir ./internal/flow ./internal/load > /tmp/array3-packages.log 2>&1
go vet ./... > /tmp/array3-vet.log 2>&1
go test ./internal/lower -count=1 -run 'TestLibraryArray' > /tmp/array3-final-lower.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -run 'TestNativeAgreesWithNode/(library_array|sort|maybe_collections|searches|adversarial_order|closures_throw)|TestArrayFamilyMutants|TestArrayWithBoundsMutant|TestArrayShrinkingSearchChecksAndMutants|TestArrayOptionalShrinkingSearchMutants' > /tmp/array3-oracle-gate.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts > /tmp/array3-counts-update.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded > /tmp/array3-counts-check.log 2>&1
```

The baseline runs the same survey against bcf803d, using a distinct work directory.
The runner is built from its own checkout, never substituted with the compiler
branch's runner. New source files in Adamic are .a. The pinned Cohere executable
cannot select .a files; exact temporary .ts mirrors use the project's strict
compiler options and prelude. They are not repository source files:

```sh
/tmp/array2-cohere --directory /tmp/array3-cohere-fixtures --types --no-fix > /tmp/array3-cohere-types.log 2>&1
/tmp/array2-cohere --no-fix --types internal/oracle/testdata/library_array_alias_is_array.a internal/oracle/testdata/library_array_sort_never.a > /tmp/array3-cohere-a.log 2>&1
/tmp/array2-cohere --directory /tmp/array3-cohere-fixtures --no-fix > /tmp/array3-cohere-full.log 2>&1
```

The ten method-family mutants are permanent in library_array_mutants_test.go and
run in TestLibraryArrayNewFamilyMutants. Each must compile and finish with exit 0,
empty stderr and no sanitizer/leak report; only Node stdout can catch it.
Two additional guard mutants were applied to lowering source separately with
restoration in finally, using the permanent negative .a fixtures:
remove the shorthand-value escape guard in libraryArrayMapReceiverUses; remove
the optional-reference partition guard in libraryArraySort. The corresponding
logs preserve the wrong native stdout versus Node. After restoring each guard,
its negative fixture again refuses. No temporary mutant test remains committed.
