package lower

import "testing"

func TestNamespaceForwarderReadinessBeforeInitialization(t *testing.T) {
	t.Parallel()
	// The mutable alias leaves preflight an unresolved call edge. Runtime must
	// reject the missing namespace before reading its uninitialized member.
	lowersAndAgreesWithNode(t, `
function read(): number { return Pending.value; }
let indirect = read;
try {
    console.log("early:" + indirect());
} catch (error) {
    if (error instanceof Error) { console.log(error.name); }
}
namespace Pending { export const value = 7; }
console.log("ready:" + indirect());
console.log("continued");
`)
}
