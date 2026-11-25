interface Item { k: number }
interface Src { a: number; next: Item | undefined }
interface Tail { a: number; next: undefined }
function isTail(s: Src): s is Tail {
    return s.a === 1;
}
interface Loose { a: number; next?: Item | undefined }
function check(value: Loose): void {
    const s = value as Src;
    if (isTail(s)) {
        console.log("tail");
    } else {
        console.log("more");
    }
}
check({ a: 1 });
check({ a: 2, next: { k: 3 } });
