# Runtime counts

Pinned stock TypeScript 6.0.3; Node v24.19.0.

Reads count getter executions after the site write and before a later store.
Violations use the declared domain, not a comparison with the old value.

| Run | Site | Assignments | Violations | Later reads | Out-of-type reads |
| --- | --- | ---: | ---: | ---: | ---: |
| fixtures | flags | 1 | 0 | 1 | 0 |
| fixtures | parent | 1 | 0 | 1 | 0 |
| flags | flags | 1 | 1 | 1 | 1 |
| flags | parent | 0 | 0 | 0 | 0 |
| parent | flags | 0 | 0 | 0 | 0 |
| parent | parent | 1 | 1 | 1 | 1 |
| acceptance | flags | 61 | 0 | 199 | 0 |
| acceptance | parent | 0 | 0 | 0 | 0 |
| compiler | flags | 171296 | 0 | 1340306 | 0 |
| compiler | parent | 0 | 0 | 0 | 0 |

The unreached fixture asserts an empty observation before the compatible controls:
both sites have zero assignments, violations and reads there.

The `fixtures` rows then test reached compatible stores. The parent compatible
store calls the observation helper directly because the upstream parent site
always stores undefined. It is a predicate control, not an upstream execution.

`inputs.csv` preserves one row per process and source filename; repeated virtual
filenames can belong to different compiler tests and are not unique case IDs.
Processes with no assignment have no CSV row. `observations.json` retains
process counts, commands, bundle hashes and sampled write/read stacks.
