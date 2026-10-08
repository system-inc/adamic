This patch is a review artifact, not applied production code.

Automatic approval review rejected shared array dispatch changes on ownership
and miscompilation risk, including backend-only staging. The owned normalizer
and disconnected selectors pass their probes and mutants; those observations
do not certify a source-level array index. This proposed patch keeps callback
and loop transfer refused and uses existing statement cleanup for an index.
It was checked with git apply --check only. See ../REPORT.md for evidence.

Optional/rest work remains a single-constructor extension: optional positions
must preserve actual source length independently of padded slots; trailing rest
positions need element contracts beyond the fixed prefix. Array-to-tuple cast
admission also requires preserving source identity across array storage. These
are pending designs, not implemented certificates.
