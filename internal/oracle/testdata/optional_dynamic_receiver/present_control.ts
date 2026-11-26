interface Options {
    readonly name?: string;
    [key: string]: string | undefined;
}
function show(value: Options): void {
    console.log(value.name ?? "missing");
}
show({ name: "set" }); show({});
