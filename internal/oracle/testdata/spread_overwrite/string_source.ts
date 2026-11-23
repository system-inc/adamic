function overwrite(source: { readonly k: string } | undefined): void {
    const copy = { ...source, k: undefined };
    console.log(String(copy.k));
}
overwrite({ k: "one" });
overwrite(undefined);
