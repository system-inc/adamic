interface Slot {
    p?: string;
    callback?: () => string;
}
const slot: Slot = {};
slot.p = undefined;
slot.callback = undefined;
function forward(value: string | undefined): Slot {
    return { p: value };
}
function consume(value: Slot): void {
    console.log(Object.keys(value).join(","));
}
consume({ p: undefined });
class Box {
    q?: number;
}
const box = new Box();
box.q = undefined;
console.log(`${"p" in slot} ${Object.hasOwn(slot, "p")} ${Object.keys(slot).join(",")}`);
console.log(`${"callback" in slot} ${slot.callback === undefined}`);
console.log(`${Object.hasOwn(forward(undefined), "p")} ${Object.hasOwn(box, "q")}`);
