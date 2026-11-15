Units newly stopped: **unknown; no valid lowering measurement**. First sites: **unknown**. Lowering coverage at current head: **0 units with aggregate-ready seeds**.

Current head: `b529db907bca6564edd2244ac8e5d102b85cf7e8`, verified descendant of `55c375d9d245b580c010c3e0b5659f2849f24fb6`.

The latent aggregate seam is stopped: independent snapshots reuse incompatible local/function IDs and discard failed-statement IR, so they do not provide faithful whole-program inputs to `findCycles`, `fresh.ProveWrites`, or the new precise written-function-capture judgment. No merger or fabricated graph is used. A static compiler-API classification is in progress, explicitly distinct from lowering measurement.

Superseded observations, including the positive and negative controls at the old head and the partial ordinary table, are retained in [REPORT-1c7f96e6-SUPERSEDED.md](REPORT-1c7f96e6-SUPERSEDED.md). No b529db90 control or root result is inferred from them.
