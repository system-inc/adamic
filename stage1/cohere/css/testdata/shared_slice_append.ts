// A dynamic owned string, then a shared slice with capacity zero.
function demonstrate(): void {
    let value = '';
    for(let i = 0; i < 256; i++) value += 'x';
    const suffix = value + value + value + value;
    value = value.slice(0, 128);
    value += suffix;
    console.log(`${value.length}`);
}
demonstrate();
