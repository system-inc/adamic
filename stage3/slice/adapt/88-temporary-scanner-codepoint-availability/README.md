Temporary: comes out when intrinsic method availability observations land

Plan published before implementation. Scanner slice only: remove the two erased
`(String as any)` casts from utf16EncodeAsStringWorker, and observe availability
with `typeof String.fromCodePoint === "function"`. Keep its conditional branches,
arrow body and utf16EncodeAsStringFallback unchanged. This answers the unchecked
cast refusal and then the unbound-method refusal of a bare availability read.

The source runs with the standard String intrinsic, which is callable when
present and undefined on older hosts. No caller replaces that intrinsic in this
proof. The typeof observation preserves both normal availability and absence;
truthy non-callable monkey patches are outside that stated input contract.
Prove all expanded Node tokens match and exercise callable/absent availability
controls. Validate the full upstream baseline before claiming the adaptation.
