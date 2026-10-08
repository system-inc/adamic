enum SymbolIndex { Absent = 0 }
function visit(index: SymbolIndex): string {
    if(index === 0) { return 'absent'; }
    return `${index}`;
}
console.log(visit(Number.parseInt('2', 10)));
