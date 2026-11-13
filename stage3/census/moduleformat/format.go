// Package moduleformat declares the bundle format used by tsc measurements.
// It has no compiler dependency, so scratch lower overlays can share the input.
package moduleformat

import (
	"fmt"
	"os"
)

func Name() string {
	value := os.Getenv("CENSUS_MODULE_FORMAT")
	switch value {
	case "", "CommonJSBundle":
		return "CommonJSBundle"
	case "EcmaScript":
		return value
	default:
		panic(fmt.Sprintf("invalid CENSUS_MODULE_FORMAT %q: want CommonJSBundle or EcmaScript", value))
	}
}

// Resolve selects the compiler's declared constants without importing the compiler
// (the scratch lower overlay itself imports this package).
func Resolve[T ~string](commonJSBundle, ecmaScript T) T {
	if Name() == "EcmaScript" {
		return ecmaScript
	}
	return commonJSBundle
}
