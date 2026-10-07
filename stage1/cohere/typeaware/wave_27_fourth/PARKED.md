# Fourth batch parked on native React analysis

Ahra now explicitly permits analysis-dependent React claims to be parked and
counted as finished for the landing-first cap. This changes claim status,
not implementation or validation results. All existing probes, independent
Go oracle, source and evidence remain pushed on codex/typeaware-wave-27.

| Claimed rule | Native dependency blocking a faithful port |
| --- | --- |
| react-hooks/set-state-in-effect | High-level IR, SSA value propagation and capture/context translation |
| react-hooks/set-state-in-render | High-level IR, SSA, post-dominance and nested setter context propagation |
| react-hooks/static-components | High-level IR, SSA phi joins and dynamic-component taint propagation; JSX parsing also missing |

All three production Go files import high_level_intermediate_representation.
set_state_in_effect.go explicitly translates nested function capture values
using Captures and Context. set_state_in_render.go relies on unconditional
blocks and nested function setter propagation. static_components.go lowers
functions, builds single-assignment form, and propagates taint through phis.
A syntax approximation would not meet the Go byte oracle.

The native analysis port is being handled on #dnv6f2c. JSX support is landing
on area/stage1-lint through integration; this worker never pushes to main or
area branches. The last fresh probe on main f8013f0b still exits 70 at the JSX
self-closing slash; Go reports the positive staticComponents finding at 74:83.
The complete commands and outputs are retained in the landing_f801 report.

No rule bodies, new bridge questions or shared files are changed by parking.
There is no new rule mutant, sanitizer, timing or byte-equivalence claim for
these three. The nine completed ports remain green and pushed on current main
f8013f0b at 0bed2f946. No source changed since that oracle revalidation.
Future selections may now proceed, excluding rules needing this native
analysis. Earlier frozen reports correctly retain their earlier pending status.
