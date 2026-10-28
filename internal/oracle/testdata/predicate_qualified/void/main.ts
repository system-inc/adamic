type Value = number | string;
namespace Debug {
export function assertNode(value: Value, test: (value: Value) => value is number): asserts value is number;
export function assertNode(value: Value, test: (value: Value) => boolean): void;
export function assertNode(value: Value, test: (value: Value) => boolean): void { console.log("called");  }
}
function use(value: Value): void { Debug.assertNode(value, (x): boolean => typeof x === "number"); console.log("after"); }
use("text");
