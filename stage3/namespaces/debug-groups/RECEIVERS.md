# Debug receivers

Exact front24 probe: native-namespace-object-receiver.a from parser-proof
5d777de3. Node and both backends print `receiver declaration loaded`. Its this
is an object method receiver, not the namespace function receiver; the old
recursive containsThis test was too strict. Arrow receiver inheritance remains.

namespace-qualified-receiver.a exercises live Debug.flags reads through this:
Node and both backends print 3 then 7. All references must be direct calls
on the declaring namespace; properties must match represented export storage
and exactly the export's declared type. No runtime object identity is invented.

namespace-detached-receiver.a prints `3:0` on Node; production refuses the
named unqualified read() call whose this is undefined. Replacing that call
with Debug.read() prints `3:3`, catching a source receiver mutant.

The compiler mutant replacing live export reads with 99 is caught by both
backend stdout comparisons. Restoring recursive scope attribution is caught
by TestNamespaceNestedReceiver. Tests pin receiver escapes, object identity
and writes through this as NotYet. Receiver proof is closed-world and forbids
unknown function-value edges rather than trusting them.

Commands: TestNamespace.*Receiver and TestNamespaceLimits in internal/lower;
TestNativeAgreesWithNode/stage3/namespaces/debug-groups in internal/oracle;
TestCountsAreRecorded -args -update-counts. Full lowered package and namespace
oracle are repeated before delivery. Source probes only declare methods, so
this does not claim Object.defineProperties or their eventual invocations.
