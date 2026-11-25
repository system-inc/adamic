interface Options { [name: string]: boolean | string | undefined }
function keep(d: { skippedOn?: keyof Options }, options: Options): boolean {
    return !d.skippedOn || !options[d.skippedOn];
}
const options: Options = { noEmit: true, 1: true, 0: true };
console.log(String(keep({ skippedOn: "noEmit" }, options)));
console.log(String(keep({ skippedOn: 1 }, options)));
console.log(String(keep({ skippedOn: 0 }, options)));
console.log(String(keep({}, options)));
