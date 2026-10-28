interface X { kind: "x"; x: number }
interface Y { kind: "y"; y: string }
type N = X | Y;
namespace ts { export function isX(value: N): value is X { console.log("called"); return false; } }
// Keep the false-direction walk live: a checked predicate also escapes as a value.
const escaped = ts.isX;
function use(n: N): void { if (!ts.isX(n)) console.log(n.y); }
use({kind: "x", x: 7});
