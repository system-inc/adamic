interface Options {
    readonly name?: string;
    [key: string]: string | undefined;
}
function show(value: Options | null): void {
    console.log(value?.name ?? "missing");
}
show(null); show({ name: "set" });
