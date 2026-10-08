function scanner(textInitial?: string): string {
    var text = textInitial!;
    const setText = (value: string): void => { text = value; };
    setText('scanner');
    return text;
}
console.log(scanner());
console.log(scanner('initial'));
let calls = 0;
function initial(): string | undefined { calls++; return undefined; }
let text = initial()!;
text = 'assigned';
console.log(`${text} ${calls}`);
function keep(value: number | undefined): number { let result = value!; return result; }
function keepBoolean(value: boolean | undefined): boolean { let result = value!; return result; }
console.log(`${keep(0)} ${keepBoolean(false)}`);
class Frame { text = initial()!; }
const frame = new Frame();
frame.text = '';
console.log(`${frame.text.length} ${calls}`);
class StaticFrame { static text = initial()!; }
StaticFrame.text = 'static';
console.log(StaticFrame.text);
function defaulted(value: string = initial()!): string { value = 'default'; return value; }
console.log(defaulted());
function assignHoisted(): void { hoistedText = 'before declaration'; }
assignHoisted();
var hoistedText = initial()!;
hoistedText = 'after declaration';
console.log(hoistedText);
