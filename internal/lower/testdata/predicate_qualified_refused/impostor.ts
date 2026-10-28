type Value = number | string;
function assertNode(value: Value, test: (value: Value) => value is number): asserts value is number;
function assertNode(value: Value, test: (value: Value) => boolean): void;
function assertNode(value: Value, test: (value: Value) => boolean): void { console.log("called");  }
const Debug: { assertNode: typeof assertNode } = {assertNode};
function use(value: Value): void { Debug.assertNode(value, (x): x is number => typeof x === "number"); }
use("text");
