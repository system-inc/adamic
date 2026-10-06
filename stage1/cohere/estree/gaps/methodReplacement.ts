class Cursor {
    next(): void {
        console.log('original');
    }
}
function replacement(): void {
    console.log('replacement');
}
const first = new Cursor();
const second = new Cursor();
first.next = replacement;
first.next();
second.next();
