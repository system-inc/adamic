# Baseline pooled-checker overflow

The TS2859 comparison of `" "` against `{}` was absent after load and first
observed by the run-end relation probe, after the final timed unit
`src/tsc/_namespaces/ts.ts:3:1`. The journal records unit duration and CPU time,
not per-unit overflow answers. Its onset unit therefore cannot be identified;
the final timed unit must not be blamed for causing it. The complete pooled
checker report and conservative all-unit attribution remain in the baseline.
