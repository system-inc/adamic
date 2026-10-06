interface IncrementInterface {
    next: (value: number) => number;
}
class Increment {
    next(value: number, step = 1): number {
        return value + step;
    }
}
function run(increment: IncrementInterface): number {
    return increment.next(4);
}
console.log(`${run(new Increment())}`);
