# Predicate reporting and census evidence

Read [REPORT.md](REPORT.md) before interpreting any count. REPORT.json retains the
complete seven-configuration finding ledger; FAMILIES.json pins exact diagnostic
messages and family counts. data/ contains compressed raw measurements; evidence/
contains tests, mutant outcomes, merge recipes and the after integration patch.

Tools are scratch-only and retain the original latent tool path. See the original
`70456b7:stage3/census/latent/README.md` for preparation and output guards. The
recorded scratch paths are provenance, not required workspace locations. Never push
a scratch census branch. The report's code configuration is the pushed feature SHA
`e58c805469e3cfa36b2d1e37b0ada2a6ae2ade6a` plus pinned original cumulative features;
this artifact commit adds evidence, without changing that compiler code.
