interface Host {
    readonly kind: "host";
    read(): string;
}
function probe(raw: { readonly kind: string }): string {
    const host = raw as Host;
    return !host.read ? "absent" : "present";
}
console.log(probe({ kind: ["ho", "st"].join("") }));
