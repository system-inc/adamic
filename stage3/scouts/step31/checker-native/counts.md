# Checker native witness counts

Local registry within this unit's territory. Programs are intentionally stopped before native execution; no native allocation/free or region counts are claimed.

| Program | Repository result | Node stdout / exit | Computed-input mutant | Native exit |
|---|---|---|---|---:|
| probes/commonjs.a | TS2591; expected-error header present | leaf / 0 | path basename changes to changed; caught | 1 |
| probes/node-require.a | NotYet: node:module.require; no header | leaf / 0 | path basename changes to changed; caught | 1 |

The stock Node checker baseline selected 301/301 acceptance projects. No central native-oracle fixtures were registered.
