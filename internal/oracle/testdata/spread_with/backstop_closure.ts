// A function through { readonly length: number }: a function's own members are Function's, which console.log's
// type shares, so a view a function reaches is one console.log reaches too, and checkLibrarySpreads refuses the
// spread at compile time, on both backends (#k7g6f9h's storage pins). A function never reaches the run-time stop.
function copy(source: { readonly length: number }): void {
    const made = { ...source, k: 1 };
    console.log(String(JSON.stringify(made)));
}
copy({ length: 2 });
copy((first: number, second: number): number => first + second);
console.log("end");
