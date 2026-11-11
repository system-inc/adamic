// The same through a pattern: each step writes item and index, and the body reassigns item (Language's
// must on #p2n17c4). Node prints 42 then 1.
class Box {
    get weight(): number {
        return 42;
    }
}
const pairs: [Box, number][] = [[new Box(), 0]];
for(let [item, index] of pairs) {
    console.log(String(item.weight + index));
    item = { weight: 1 };
    console.log(String(item.weight + index));
}
