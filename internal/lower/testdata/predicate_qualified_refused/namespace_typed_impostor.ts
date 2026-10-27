type Value = number | string;
namespace Real {
export function assertNode(value: Value, test: (value: Value) => value is number): asserts value is number;
export function assertNode(value: Value, test: (value: Value) => boolean): void;
export function assertNode(value: Value, test: (value: Value) => boolean): void { console.log("called");  }
}
const Debug: typeof Real = { assertNode: Real.assertNode };
function use(value: Value): void { Debug.assertNode(value, (x): x is number => typeof x === "number"); }
use("text");
