Commands, on compiler/no-submodule-recursion-main 4a1149f3 (base origin/main 578c0b3d), git 2.43.0, Go 1.27.1, as root, TMPDIR=/tmp/adamic-gate:
  go vet ./internal/buildcache/   (clean)
  go test -race -count=1 ./internal/buildcache/ > evidence/buildcache.log   (exit 1: the 5 root-only failures)
  the same on origin/main > evidence/buildcache-main-578c0b3d.log   (exit 1: the same 5; *.failing lists equal)
Mutants: evidence/mutants/<name>.diff against the change, <name>.log the run of -run 'Submodule|TreeState|TreeThatChanged';
m4 and m5 also ran the whole package (<name>.whole-package.log): only the 5 root-only failures.
  rerun on the committed dde5ea8c: evidence/buildcache-dde5ea8c.log (the same 5 failures as main)
