interface Options {
    name: string;
    [key: string]: string | number | undefined;
}
interface Index {
    [key: string]: string | number | undefined;
}
function erase(index: Index, key: string): void {
    delete index[key];
}
function show(value: Options): void {
    console.log(value?.name ?? "missing");
    try {
        console.log(String(value?.name.length));
    } catch {
        console.log("caught");
    }
}
const value: Options = { name: "set" };
erase(value, "name");
show(value);
