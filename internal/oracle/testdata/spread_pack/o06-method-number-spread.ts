type A = number | string;
const nums: number[] = [4, 5];
class Shelf {
    put(...rest: A[]): number { return rest.length; }
}
console.log(String(new Shelf().put(...nums)));
