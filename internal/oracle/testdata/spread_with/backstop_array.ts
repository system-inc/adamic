// An array through { readonly length: number }: a library type in the closed world the spread's view may hold, so
// checkLibrarySpreads refuses the spread at compile time, on both backends (#k7g6f9h's storage pins).
function copy(source: { readonly length: number }): void {
    const made = { ...source, k: 1 };
    console.log(JSON.stringify(made));
}
copy({ length: 2 });
copy([1, 2]);
