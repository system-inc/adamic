// A for-of's let binding, reassigned in the loop's body: the first read is the element, a Box whose
// weight is a getter, and the second the literal (Language's must on #p2n17c4). Node prints 42 then 1.
class Box {
    get weight(): number {
        return 42;
    }
}
const items: Box[] = [new Box()];
for(let item of items) {
    console.log(String(item.weight));
    item = { weight: 1 };
    console.log(String(item.weight));
}
