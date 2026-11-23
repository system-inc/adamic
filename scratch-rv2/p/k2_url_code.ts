const own = Object.prototype.hasOwnProperty;
try {
    new URL("not a url");
} catch (caught) {
    if (caught instanceof Error) {
        console.log(String(own.call(caught, "code")));
    }
}
