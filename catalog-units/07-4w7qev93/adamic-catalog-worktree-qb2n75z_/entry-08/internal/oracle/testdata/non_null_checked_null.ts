function show(value: RegExpExecArray | null): void {
    console.log('before');
    console.log(String(value! === null));
}
show(null);
