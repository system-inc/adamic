interface Options {
    name?: string;
    [key: string]: string | number | undefined;
}
interface Index {
    [key: string]: string | number | undefined;
}
function store(index: Index, key: string): void {
    index[key] = 3;
}
function show(value: Options | undefined): void {
    console.log(value?.name ?? "missing");
}
const value: Options = {};
store(value, "name");
show(value);
