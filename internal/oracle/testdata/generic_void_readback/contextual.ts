type Result<T> = readonly T[] | "missing";
function collect<T>(found: boolean): Result<T> {
    const items: T[] = [];
    if (found) return items;
    return "missing";
}
const result: Result<number> = collect(true);
console.log(typeof result === "string" ? result : String(result.length));
