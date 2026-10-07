function read(): string {
    let value: string | undefined;
    return value === undefined ? 'missing' : 'present';
}
console.log(read());
