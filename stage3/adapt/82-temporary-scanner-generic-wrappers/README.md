Temporary: comes out when generic callback function widening lands

Plan published before implementation. Slice scanner.ts only: the returned
scanner object's tryScan, lookAhead and scanRange properties use explicit
generic forwarding arrows with the same arguments and return values. Keep
Scanner's interface and original implementations unchanged. Merely changing
interface method signatures to properties did not close the refusal; the
contextual forwarding arrows do. Census reason: method-signature-style.

Validate tokens, callback results/state restoration and full upstream baseline.
The wrappers add a frame and a function object; scanner's original helpers are
private and use no this. Function source reflection is outside this token proof.

Validated with 59/80-83: 106,367 upstream tests pass, zero failures/pending/
differences, 230.719 seconds. All Node tokens match. Only the three object
properties change; the public Scanner declarations are retained. The minimal
wrapper probe advances to scanner:949:48, a zero-valued EscapeSequenceScanningFlags
ternary branch losing its flag domain. Property signatures alone did not help.
