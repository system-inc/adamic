// A Map through { readonly size: number }: a library type in the closed world the spread's view may hold, so
// checkLibrarySpreads refuses the spread at compile time, on both backends (#k7g6f9h's storage pins).
function copy(source: { readonly size: number }): void {
    const made = { ...source, k: 1 };
    console.log(String(JSON.stringify(made)));
}
copy({ size: 2 });
copy(new Map<string, number>([["a", 1]]));
