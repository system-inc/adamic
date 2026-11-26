interface Options {
    readonly name: string;
    [key: string]: string | undefined;
}

function show(value: Options | undefined): void {
    const length = value?.name.length;
    console.log(length === undefined ? "undefined" : String(length));
}

show(undefined);
show({ name: "set" });
