function overwrite(source: { readonly k: number }): void {
    const copy = { ...source, k: undefined };
    console.log(String(copy.k));
}
overwrite({ k: 1 });
