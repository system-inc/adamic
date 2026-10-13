Ported the speculative census onto main without importing the old base.
Source commits: e46fa73d, 20ebb633 and f6bb0b41; historical tables remain labeled by their old base.
Focused depth, topology, binding, snapshot, output isolation and header controls pass on main.
Their mutants fail the intended assertions; normal C and JS stay byte-identical to main.
Whole-project streaming and its bounded observations are documented in streaming.md and README.md.

Base: `946a8f095a7fa419a92117406314b7b3d44630f0`. Applying the three commits with
`git cherry-pick --no-commit` succeeded without conflicts. No stage1 file was carried.
Main-side latent replay, refusal-rewriter and copier tests were retained.

The old base provided `signatureReturn`; main computes returns in `signature`.
The scratch overlay adapts that entry point. Main's opaque checker mapper facade
must retain pointer identity, so the census state copier shares it. Main's IR has
a private backend argument-facts cache; copying accepts only its nil state and
fails on a populated cache. Main supports debugger, so nesting witnesses now use
three nested with statements and a sibling with. Their eight sites retain exact
0/1/2/0 depths. The imported-body witness likewise uses a nested with.

Validation commands used `/workspace/adamic-tools/env.sh`, with every invocation
under timeout and every stdout/stderr stream in a named log. No package-wide
confirmation or full gate was run.

- `make_overlay.py` and `go build -buildvcs=false -overlay=... ./stage3/census/latent/tool`.
- `audit.py`: eight sites; depths 0/1/2/0; full/no-stubs exact; no-stubs and depth-zero caught.
- `audit_signature.py`: five checker-clean sites, depth 3; both mutants caught.
- `audit_dependency.py`: foreign depths 1/2; depth-zero caught.
- `audit_topology.py`: eight sites and stock ancestry; omitted-boundary, legacy-site-key and span-only caught.
- `prove_identity.py`: token attribution mutant caught.
- `audit_snapshot.py`: baseline passes; shared slice, scalar values, cursor, private field, mapper identity and schema mutants caught. Every Go test uses `-timeout 90s`.
- `audit_binding.py`: hidden binding bypass caught.
- `isolation.py`: main, overlay off and production speculative flag are identical.
  C: 6,907 bytes, SHA-256 `169abd45ca4c225362f18da3c6c8bd8171a9939f51e329b2686e7152c77479ab`.
  JS: 9,930 bytes, SHA-256 `918453f11ba043e10346c3dd920e5fba93d72ccca6aaed92f7126c49a808f7f5`.
  Both appended-output mutants are caught.
- Unchanged Gate.aCheck at `0b76ca8b88a0ddfc9c44d17666b80bc4cbd93b40`: six files, zero mismatches, eight missing/wrong-code mutants caught.

Setup: first `timeout 240 bash cloud/setup.sh` reached its deadline during cold
build warming. Earlier timing lines were Go 0.062s, Node 0.087s, clang 0.504s,
markdown 1.191s, submodules 17.828s, shared cache 23.702s. `nproc` is 5;
the box's CPU quota is 4. A bounded warm retry is logged separately.

The dispatched mapper fixture lowers successfully on main without a production
change. Source Node, generated JS Node and sanitized native output match. Reverting
main inference commit 6b33ec61 in a scratch overlay fails the fixture. No production
mapper commit is needed; streaming.md records the equivalent main commit.
