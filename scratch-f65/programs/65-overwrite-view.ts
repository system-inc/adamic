function overwrite(source: { readonly k?: undefined }) {
    const copy = { k: 1, ...source };
    console.log(String("k" in copy));
    console.log(String(copy.k));
}
overwrite({ k: undefined });
