function classify(unit: number): number {
    switch(unit) {
        case -1:
            return 1;
        default:
            return 0;
    }
}
console.log(String(classify(-1)));
