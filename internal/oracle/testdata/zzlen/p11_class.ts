class Pair {
    readonly left: number;
    readonly right: number;
    constructor(left: number, right: number) { this.left = left; this.right = right; }
}
function arity(f: Function): number { return f.length; }
console.log(String(arity(Pair)));
