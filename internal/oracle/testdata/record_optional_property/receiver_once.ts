interface Options {
    readonly name?: string;
    [key: string]: string | undefined;
}

let count = 0;
function receive(value: Options | undefined): Options | undefined {
    count = count + 1;
    return value;
}

const value: Options = { name: "set" };
console.log(receive(value)?.name ?? "missing");
console.log(String(count));
console.log(receive(undefined)?.name ?? "missing");
console.log(String(count));
