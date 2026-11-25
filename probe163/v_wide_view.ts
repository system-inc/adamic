interface Item { k: number }
interface Src { a: number; next: Item | undefined }
interface Tail { a: number; next: undefined }
function isTail(s: Src): s is Tail {
    return s.a === 1;
}
interface Wide { a: number }
const wide: Wide = { a: 1 };
const s = wide as Src;
console.log(isTail(s) ? "tail" : "more");
