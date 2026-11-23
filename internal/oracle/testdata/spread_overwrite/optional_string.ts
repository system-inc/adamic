function overwrite(source: { readonly k?: string }): void {
    const copy = { ...source, k: undefined };
    console.log(String("k" in copy));
    console.log(String(copy.k));
}
overwrite({ k: "one" });
overwrite({});
