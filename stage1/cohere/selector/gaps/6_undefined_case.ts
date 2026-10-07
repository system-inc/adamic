function letter(value: string | undefined): string {
    switch(value) {
        case undefined:
            return '?';
        default:
            return value;
    }
}
console.log(letter('a'));
