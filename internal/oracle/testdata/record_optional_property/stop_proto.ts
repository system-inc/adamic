interface Options {
    readonly name?: string;
    [key: string]: string | undefined;
}

function show(value: Options | undefined): void {
    console.log(typeof value?.__proto__);
}

show({});
