# Mutants for cloud/env-owner.sh (#zb2bhp9)

Run `bash cloud/env-owner-test.sh` with each edit, then put the file back.

| Mutant | Edit | Red |
|---|---|---|
| The writer's refusal removed | cloud/write-env.sh: the `adamic_env_refuse_release_owned "$environment"` line deleted | "write-env.sh refuses a marked env.sh": it exits 0 and rewrites the release's file |
| setup.sh's refusal removed | cloud/setup.sh: the `adamic_env_refuse_release_owned "$tools/env.sh"` line deleted | "setup.sh stops before taking its lock or installing": setup takes $tools/setup.lock and installs into the release's tool directory before write-env.sh's own refusal stops it, still with exit 2, so the lock file is the check that tells the two apart |
| The marker matched loosely | cloud/env-owner.sh: `grep -qxF` made `grep -qF "adamic-env"` | none here; it's a looser match that still refuses, kept as an equivalent. The exact line is Loom's contract |
