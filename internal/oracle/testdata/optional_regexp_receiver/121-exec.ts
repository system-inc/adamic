function check(regex: RegExp | undefined, text: string): void {
    const match = regex?.exec(text);
    console.log(typeof match);
    console.log(String(match?.length));
}
check(/entry/, "entry");
check(/entry/, "other");
check(undefined, "entry");
