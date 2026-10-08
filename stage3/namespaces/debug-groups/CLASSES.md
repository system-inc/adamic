# Debug namespace class

The exact native-namespace-class.a probe from parser-proof 5d777de3 now prints
`class declaration loaded` in Node, native and emitted JavaScript.
namespace-class-initialization.a adds static and per-instance side effects,
constructor effects, distinct instances, live static reads and instanceof.
The original declaration-shaped class NotYet is removed; existing class
representations and independent class restrictions remain.

namespace-class-unready.a calls a private helper before the class declaration:
all backends stop at ReferenceError: Cannot access 'DebugTypeMapper' before
initialization, exit 70, without reaching the final print.
namespace-class-unknown-before.a reaches construction through a mutable
callable before namespace initialization: all backends stop at TypeError:
Cannot read properties of undefined (reading 'DebugTypeMapper'), exit 70.

Mutants, both restored: removing namespace-class readiness storage makes the
early class fixture print unreachable with exit 0, caught by both backend
comparisons. Removing the constructor's outer namespace check produces the
wrong ReferenceError instead of TypeError, caught by stderr in both backends.
No mutant passed through a Go compilation failure.

Commands: go test ./internal/lower; focused TestNativeAgreesWithNode for the
four class fixtures; namespace-wide TestNativeAgreesWithNode and TestNamespace;
TestCountsAreRecorded -args -update-counts. Counts refreshed. The normalized
Debug shape advances to callable namespace merging. The original twelve real
namespace fixtures retain six Compiles, three NotYet and three Refused.
