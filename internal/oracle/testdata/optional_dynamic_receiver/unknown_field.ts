interface Holder {
    readonly data: unknown;
}
function show(holder: Holder | null): void {
    const data = holder?.data;
    console.log(typeof data === "string" ? data : "missing");
}
show(null); show({ data: "set" });
