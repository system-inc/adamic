function match(type: string, indicator: string): string {
    switch(type) {
        case indicator: return '1';
        default: return '0';
    }
}
console.log(match('doc-start', 'doc-start'));
