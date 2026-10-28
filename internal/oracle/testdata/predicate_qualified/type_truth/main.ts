type Value = number | string;
namespace Debug { export function type<T>(value: unknown): asserts value is T; export function type(value: unknown): void { console.log("called"); } }
function use(value: Value): void { Debug.type<number>(value); console.log("after"); }
use(7);
