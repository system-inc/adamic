# Where the daily meter's 14.85 GiB goes (task #rw2vjhn) - INTERIM

Status: step 1 is partly done. This file is updated as steps finish. Measured and inferred are kept apart.

## Box (measured)
Intel Xeon @ 2.10GHz, 4 cores, 15 GiB RAM (free -m: 16091 MB total), no swap. No cgroup memory.max file visible in this container (so no cgroup limit to read; the box itself is the limit). Go 1.27.1, Node v22.23.4 (not 24), GNU time 1.9.
Compiler sha: origin/main 9d850b4f6bb10cfedd2d6ad747972833947d2cbe (the Nov 4 run was f87d78848071; I did not use it). Upstream TypeScript v6.0.3 050880ce. Built in "single" mode as build_census does; tree from stage3/apply.sh.
Full mode (LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1) as the brief says. Note: stage3/meter/README.md's daily command sets STAGE3_METER_LATENT=first-error, while the script default is full. I did not establish which the Nov 4 run used.

## Step 1 so far (measured, /usr/bin/time -v, nothing else running)
| step | max RSS | wall | state |
| --- | ---: | ---: | --- |
| 1 census on tsc.ts | 868,036 KB (0.83 GiB) | 23.4 s | done, exit 0 (log: 1-census.time) |
| 2 entry-census (full) on tsc.ts | still running | >90 min | RSS sampled with ps: 1.21 GB at 8 min, 1.39 GB at 18 min, 1.68 GB at 30 min, 1.78 GB at 39 min, 2.53 GB at 60 min, 2.85 GB at 90 min |
| 3 census on src/compiler | not yet run | | |
| 4 latent-census (full) on src/compiler | not yet run | | |

## Reading so far (inferred)
Step 2 grows steadily, not as a plateau, but at 90 minutes it is 2.85 GB, nowhere near 14.85 GiB. Either the 14.85 GiB peak is step 4 (or a later part of step 2), or it comes from a different mode than I am running. Not yet established.
