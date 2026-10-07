Status: appendSuccessor is claimed but blocked by strong recursive CFG successor ownership; no helper delivered.
SHAs: claim 268ad5b63; existing helper re-green efc11a885; rules parked c74d385b6 for #zmh9v36.
Checks: actual private Go appendSuccessor and source Node both exit 0 with empty stderr and output 1,7,true. Both Adamic backends stop in lower.Lower with adamic/cycle-capable before emission.
Mutant: source target-index mutation exits 0 with empty stderr and outputs 1,8,true; only Go comparison catches it. No native or emitted-JavaScript mutant catch is claimed.
Uncovered: four consumer corpora, duplicate edges, Go slice backing aliases and native/emitted output parity; zero rules unblocked, no subsequent claim.

The exact failure is retained in evidence/successor-gap/build.log. The helper
writes a supplied block into a supplied block's strong recursive successor array.
A self-edge is valid Go CFG behavior, confirmed through an overlay adapter that
calls the private upstream method; it is refused by Adamic's ownership check.
Go and source Node observations are retained separately from that compiler refusal.
No shared harness or compiler file changed. The candidate stays under gaps/,
not exported as a helper, and readiness is not increased.

The diagnostic suggests weak references. A weak successor graph requires an
explicit owning block arena and a matching graph API, including undefined reads
after ownership ends; it cannot silently replace the existing strong Block[]
contract. That representation decision must be coordinated with the CFG block
helper owner before this port can be certified. Go slice header/backing alias
semantics are also unproved and must not be hidden by an array-push approximation.

The current primitive result is enough to block complete four-consumer parity,
so no whole-rule corpus or throughput claim is made. Existing Theme.Add and
FrameworkStaticReading remain independently green on current main c01907a7.
