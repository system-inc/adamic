interface Options {
    readonly label: string;
    [key: string]: string | undefined;
}
function size(value: Options | null | undefined): void {
    console.log(String(value?.label.length));
}
size(null);
size({ label: "abc" });
size(undefined);
