// k is at slot 0 in a plain literal, and appended at slot 1 by a spread over { a }, so a read of k goes by name
// (#k7g6f9h).
const plain = { k: 10, a: 0 };
function add(source: { readonly a: number }): { a: number; k: number } {
    return { ...source, k: 7 };
}
const made = add({ a: 1 });
console.log(`${plain.k} ${plain.a} ${made.k} ${made.a}`);
