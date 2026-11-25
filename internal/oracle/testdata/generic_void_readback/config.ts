interface ProjectReference { readonly path: string; }
type PropOfRaw<T> = readonly T[] | "not-array" | "no-prop";
function getPropFromRaw<T>(prop: "files" | "include" | "exclude" | "references",
    validateElement: (value: unknown) => boolean, elementTypeName: string): PropOfRaw<T> {
    if (prop !== "references") return "no-prop";
    if (!validateElement(elementTypeName)) return "not-array";
    const result: T[] = [];
    return result;
}
const raw = getPropFromRaw<ProjectReference>("references", value => typeof value === "string", "object");
console.log(typeof raw === "string" ? raw : String(raw.length));
