function spread(source: { readonly k: string } | undefined) {
    return { ...source };
}
function overwrite(source: { readonly k?: string }): void {
    const copy = { ...source, k: undefined };
    console.log(String(copy.k));
}
overwrite(spread({ k: "one" }));
overwrite(spread(undefined));
