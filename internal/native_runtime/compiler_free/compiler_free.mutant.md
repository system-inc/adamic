# The runtime imports no compiler package: a mutant (#tb722h2 H2)

The mutant is one edit, applied alone, then reverted. Run
`go test ./internal/native_runtime/compiler_free -run '^TestTheRuntimeImportsNoCompilerPackage$' -count=1 -v`.
Not run yet: written on a machine that doesn't build Go, so "goes red" below is what the code says, not what a run
printed.

## M1. A blank import of internal/ir in the runtime

File: internal/native_runtime/options.go, its import.

Old:

```go
import "os"
```

New:

```go
import (
	"os"

	_ "github.com/system-inc/adamic/internal/ir"
)
```

Expected: TestTheRuntimeImportsNoCompilerPackage goes red with `internal/native_runtime imports the compiler, so every
compiler edit moves every runtime archive's key: github.com/system-inc/adamic/internal/ir`, followed by any other
package of compilerPackages, or under one, that internal/ir imports in turn.
