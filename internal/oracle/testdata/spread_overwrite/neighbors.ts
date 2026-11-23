function spread(source: { readonly k: number } | undefined): void {
    const copy = { ...source };
    console.log(String(copy.k));
}
function overwrite(source: { readonly k: number } | undefined): void {
    const copy = { ...source, k: 2 };
    console.log(String(copy.k));
}
function bump(source: { readonly k: number; readonly label: string }): void {
    const copy = { ...source, k: source.k + 1 };
    console.log(`${copy.label} ${copy.k}`);
}
spread({ k: 1 });
spread(undefined);
overwrite({ k: 1 });
overwrite(undefined);
bump({ k: 1, label: "one" });
