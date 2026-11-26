interface Holder {
    readonly data: unknown;
}
function show(holder: Holder | null | undefined): void {
    console.log(holder?.data ?? "missing");
}
show(null); show({ data: "set" }); show(undefined);
