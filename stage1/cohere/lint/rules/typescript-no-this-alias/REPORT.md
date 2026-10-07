The listener, exact messages, descriptor, independent Go options adapter,
witnesses and initializer-unwrapping mutant are implemented in this directory.
All 40 original configurations and every compiler/stage1 source match Go on
Node source, emitted JavaScript and sanitized native. The compiling mutant is
caught on all three paths. The whole source corpus has zero aliases; positive
fixtures independently prove that the rule reports them.

See [the complete evidence and throughput](../typescript-no-non-null-asserted-optional-chain/REPORT.md).
This rule has no suggestions or fixes. The shared lint package remains blocked by
its existing profile_test.go compilation failure. Sources use Ahra's authorized
.ts fallback. No shared file was changed.
