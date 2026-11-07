# Construction proof checkpoint

Both full applies pass on the frozen main plus getownvalues stack. The accepted
full apply has exactly the 746 source files used in the measured generic-helper
census, byte for byte. Only hasProperty's pinned header differs from baseline;
assign and every caller remain byte-identical.

The ordinary compiler builds before and after pass. Adamic types over all 79
sorted compiler roots passes before and after, with zero diagnostics. Both
public files, typescript.d.ts and tsserverlibrary.d.ts, are byte-identical.
All nine emitted JavaScript files are byte-identical. Complete core.ts emission
is byte-identical with comments kept and removed. typescript.internal.d.ts
changes exactly the hasProperty generic signature and nothing else.

The original-helper mutant restores all 746 baseline source bytes. Its 79-root
type check passes and ordinary binder census returns NotYet core.ts:1364:29,
a non-object representation passed through an opaque object parameter.

The shared owner contract passes 12 no-write refusals, foreign-edit preservation,
LF/CRLF preservation and zero-write, byte/mtime-preserving replay. The frozen
baseline replay completed all 83 adapters, with its 14 inherited nonzero exits
and zero changed source bytes. The accepted replay is running; its paired
comparison and final archival/cleanup remain pending.

comparison.json and AFTER-PHASES.json retain these completed results. The final
proof will replace this checkpoint once the accepted replay completes.
