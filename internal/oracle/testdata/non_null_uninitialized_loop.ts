function run(flag: boolean): void {
    let value: number = null!;
    while (flag) { value = 7; flag = false; }
    console.log(`${value}`);
}
run(false);
