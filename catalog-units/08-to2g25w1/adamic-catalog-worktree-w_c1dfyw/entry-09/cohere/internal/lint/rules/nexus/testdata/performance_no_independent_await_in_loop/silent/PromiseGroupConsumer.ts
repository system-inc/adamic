// Shape from ahra libraries/structure/libraries/nexus/source/coordination/PromiseGroup.test.ts:304.
// The promises were started before the loop; awaiting each in turn serializes no work, so
// Promise.all would change only the spelling. Silent because every await's operand is the item.
interface PromiseGroupInterface<ValueType> {
    add(promise: Promise<ValueType>): void;
    next(): Promise<ValueType>;
}
declare function createPromiseGroup<ValueType>(): PromiseGroupInterface<ValueType>;

export async function consumeAll(): Promise<number[]> {
    const group = createPromiseGroup<number>();
    const results: number[] = [];

    for(let value = 1; value <= 5; value++) {
        group.add(Promise.resolve(value));
    }

    const nextPromises = [];
    for(let iteration = 0; iteration < 5; iteration++) {
        nextPromises.push(group.next());
    }

    for(const promise of nextPromises) {
        const value = await promise;
        results.push(value);
    }

    return results.sort();
}
