function show(value: number | undefined): void {
    console.log('before');
    let stored: number = value!;
    console.log(String(stored));
}
show(undefined);
