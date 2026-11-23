const own = Object.prototype.hasOwnProperty;
try {
    const size: number = -1;
    console.log(String(new Array<number>(size).length));
} catch (caught) {
    if (caught instanceof Error) {
        console.log(String(own.call(caught, "code")));
    }
}
