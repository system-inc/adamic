function check(text: string, args: [RegExp]): void {
    const value = text?.match(...args);
    console.log(String(value?.length));
}
check("entry", [/entry/]);
