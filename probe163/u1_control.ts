interface Item { k: number }
interface Src { a: number; next: Item | undefined }
interface Tail { a: number; next: undefined }
function isTail(s: Src): s is Tail {
    return s.next === undefined;
}
const s: Src = { a: 1, next: undefined };
const t: Src = { a: 2, next: { k: 3 } };
console.log(isTail(s) ? "tail" : "more");
console.log(isTail(t) ? "tail" : "more");
