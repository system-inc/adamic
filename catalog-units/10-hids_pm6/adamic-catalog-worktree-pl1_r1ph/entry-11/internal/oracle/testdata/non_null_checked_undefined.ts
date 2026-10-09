function show(value: string | undefined): void {
    console.log('before');
    console.log(String(value! === undefined));
}
show(undefined);
