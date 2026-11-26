// A Set through { readonly size: number }: lowering stops it where it's passed, before the spread, since a Map or a
// Set held where a plain object type is held isn't lowered yet (#k7g6f9h's storage pins). checkLibrarySpreads would
// refuse the spread too, so a Set never reaches the run-time stop.
function copy(source: { readonly size: number }): void {
    const made = { ...source, k: 1 };
    console.log(String(JSON.stringify(made)));
}
copy({ size: 2 });
copy(new Set<number>([1]));
