function overwrite(source: { readonly k: number } | undefined): void {
    const copy = { ...source, k: undefined };
    console.log(String(copy.k));
}
overwrite({ k: 1 });
overwrite(undefined);
