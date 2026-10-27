type Value = number | string;
export namespace Debug {
export function assertNode(value: Value, test: (value: Value) => value is number): asserts value is number;
export function assertNode(value: Value, test: (value: Value) => boolean): void;
export function assertNode(value: Value, test: (value: Value) => boolean): void { console.log("called"); if (!test(value)) throw new Error("bad input"); }
}
