The main port's controls and their mutants pass; the whole-project census remains bounded and resumable.
Compiler base: 946a8f095a7fa419a92117406314b7b3d44630f0; original corpus bytes are preserved.
Every measurement and build has an outer timeout; focused Go tests use -timeout 90s.
All outputs are captured in named logs, losslessly archived under evidence/main-stream/.
Historical validation and directory evidence are retained at f6bb0b41 and evidence/directories/.

Source `/workspace/adamic-tools/env.sh` after setup, with GOPROXY exported as
`https://proxy.golang.org|direct`. The cold setup hit its 240s limit during build
warming. The warm retry completed in 74.816s: Go/node .079s, submodules .120s,
markdown .122s, clang .201s, shared cache .500s, build 74.698s, test binaries
74.788s, cache warm 74.789s. nproc=5, cgroup quota=4 CPUs. Go 1.27.1,
Node 24.19.0, clang 20.1.8. No whole-package confirmation or full gate was run.

Focused logged checks: audit.py (eight nested sites), audit_signature.py (five
checker-clean sites through depth 3), audit_dependency.py (foreign depths 1/2),
audit_topology.py (eight exact identities), prove_identity.py, audit_snapshot.py,
audit_binding.py, audit_output_guards.py and isolation.py. They reject depth-zero,
no-stubs, omitted-boundary, legacy-site-key, span-only, token-identity, hidden-binding,
shared-slice, scalar-value, cursor, private-field, mapper-identity, schema,
IR-output and loader-output mutants. The unchanged Gate.aCheck at 0b76ca8b
checks six controls with zero mismatches and catches eight header mutants.

The final streaming binary passes audit_stream.py: buffered findings parity,
resume skipping, checksum rejection, dropped successful records, depth shifted by
one and independent arithmetic mutants. Resource witnesses catch a sleeping
process at its deadline, resident-memory excess and address-space allocation
failure. RLIMIT_AS is the hard memory cap; RSS watchdog sampling is documented.

The standalone checker probe uses timeout 90s, 6GiB RSS allowance and 12GiB hard
address-space cap. It exits 124 after 90.099s with peak RSS 940780KiB and no complete
record. All-file speculative measurements use 30s per file under an outer 3600s
timeout. Corpus full/no-stubs baselines use 5s per file and count only completed
records. No 15-minute interval is left unmonitored.

The mapper fixture from a64f77ee has matching source Node, generated JS Node and
sanitized native output, with empty stderr. A scratch overlay reverting inferTypes
from 6b33ec61 to its parent fails that fixture. Main's equivalent fix is therefore
reported separately as already on main; no production mapper commit is carried.

Pinned main, normal overlay-off and speculative flag on the production binary
emit identical C (6907 bytes) and JS (9930 bytes). Appended-output mutants fail
both byte comparisons. Production source diff is empty.

Reproduction uses the commands in streaming.md. Run finalize_stream.py under
an outer timeout with NODE_PATH selecting the pinned 6.0.3 stage3 API cache, then
verify.py on the resulting raw stream, RESULT.json and compiler root. Finalizer
and verifier log the required shifted-depth and dropped-file failures. Lane
checks run after each commit before its push; their logs are archived.
