namespace Debug { export function assertIsDefined<T>(value: T): asserts value is NonNullable<T> { console.log("called"); if (value === undefined) throw new Error("bad input"); } }
function use(value: number | undefined): void { Debug.assertIsDefined(value); console.log("after"); }
use(7);
