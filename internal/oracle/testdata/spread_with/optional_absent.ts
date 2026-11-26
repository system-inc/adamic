// An optional field the actual source lacks is appended, and one it holds keeps its place (#k7g6f9h).
function fill(source: { readonly a: number; readonly k?: string; readonly z: number }): void {
    const copy = { ...source, k: "set" };
    console.log(Object.keys(copy).join(","));
    console.log(JSON.stringify(copy));
    console.log(`${copy.a} ${copy.k} ${copy.z}`);
}
fill({ a: 1, z: 3 });
fill({ a: 1, k: "old", z: 3 });
