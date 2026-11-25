interface Slot { pattern: RegExp | undefined }
const slot: Slot = { pattern: /entry/ };
function clear(): void { slot.pattern = undefined; }
function input(): string { console.log("evaluated"); return "entry"; }
if (slot.pattern !== undefined) {
    clear();
    console.log(String((slot.pattern?.test)(input())));
}
