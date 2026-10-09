function show(value: number | undefined): void {
    console.log('before');
    var stored: number = value!;
    console.log(String(stored));
}
show(undefined);
