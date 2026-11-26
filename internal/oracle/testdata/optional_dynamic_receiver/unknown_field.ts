interface Holder {
    readonly data: unknown;
}
function show(holder: Holder | null): void {
    console.log(holder?.data ?? "missing");
}
show(null); show({ data: "set" });
