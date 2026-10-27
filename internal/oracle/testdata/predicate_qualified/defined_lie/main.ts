namespace Debug { export function assertIsDefined<T>(value: T): asserts value is NonNullable<T> { console.log("called");  } }
function use(value: number | undefined): void { Debug.assertIsDefined(value); console.log("after"); }
use(undefined);
