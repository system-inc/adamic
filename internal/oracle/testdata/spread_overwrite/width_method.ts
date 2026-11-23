function overwrite(source: { readonly a: number }): void {
    const copy = { ...source, k() { return 5; } };
    console.log(`${copy.a} ${copy.k()}`);
}
const wide = { a: 1, k: "x" };
overwrite(wide);
