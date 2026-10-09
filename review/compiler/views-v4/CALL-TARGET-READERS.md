# Callable view reader routing

The two lowering readers now obtain the invoked value from ClosureTargets. Its
Value preserves receiver flow; Unknown still governs the function target set and
is never converted into a proof. Direct-call emission alone has a compiler-owned
allowlist entry, with the reason "names the function to call".

Validation: focused IR targets and TestCallTargetReaders passed (23.333 seconds);
focused lower view tests passed (0.224 seconds). No runtime checks changed and no
new fixture or counts rows were introduced. No mutant is claimed for this routing
change. Setup passed in 9.964 seconds; nproc 5, CPU quota 4.
