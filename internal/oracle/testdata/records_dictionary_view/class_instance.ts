// A class instance handed on as object is a fixed object of its class. Node reads its name by key; both backends stop
// at the cast to MapLike<unknown>, a run-time NotYet, after the line before it.
interface MapLike<T> {
    [index: string]: T;
}
class Package {
    readonly name: string;
    constructor(name: string) {
        this.name = name;
    }
}
function packageName(exports: object): unknown {
    return (exports as MapLike<unknown>)["name"];
}

console.log("before");
console.log(typeof packageName(new Package("class")));
console.log("after");
