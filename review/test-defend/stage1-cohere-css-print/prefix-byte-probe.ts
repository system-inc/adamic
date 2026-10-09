function probe(): void {
    let value = '';
    for(let index = 0; index < 256; index++) value += 'x';
    const suffix = value + value + value + value;
    value = value.slice(0, 128);
    value += suffix;
    console.log(`${value.length} ${value.charCodeAt(127)}`);
}
probe();
