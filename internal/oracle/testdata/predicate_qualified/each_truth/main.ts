interface NumberNode { kind: "n"; value: number }
interface StringNode { kind: "s"; text: string }
type Value = NumberNode | StringNode;
namespace Debug {
 export function assertEachNode(values: readonly Value[], test: (value: Value) => value is NumberNode): asserts values is readonly NumberNode[];
 export function assertEachNode(values: readonly Value[], test: (value: Value) => boolean): void;
 export function assertEachNode(values: readonly Value[], test: (value: Value) => boolean): void { console.log("called"); for (const value of values) if (!test(value)) throw new Error("bad input"); }
}
function use(values: readonly Value[]): void { Debug.assertEachNode(values, (x): x is NumberNode => x.kind === "n"); console.log("after"); }
use([{kind: "n", value: 7}]);
