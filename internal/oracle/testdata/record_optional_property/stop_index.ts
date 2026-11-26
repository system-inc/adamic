interface Options {
    readonly name?: string;
    [key: string]: string | undefined;
}

function show(value: Options, key: string): void {
    console.log(value?.[key] ?? "missing");
}

show({}, "name");
