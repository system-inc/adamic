interface ReadBase {
    readonly name?: string;
    readonly kind: number;
}
interface WritableBase {
    name: string;
    readonly kind: number;
}
interface Arrow extends ReadBase {
    readonly name: never;
}
function makeArrow(made: ReadBase): Arrow {
    return made as Arrow;
}
const base: WritableBase = { kind: 1, name: "before" };
const arrow = makeArrow(base);
console.log(String(Object.is(base, arrow)) + " " + String(arrow.kind) + " " + base.name);
base.name = "written";
console.log(String(Object.is(base, arrow)) + " " + String(arrow.kind) + " " + base.name);
