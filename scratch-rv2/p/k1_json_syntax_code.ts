const own = Object.prototype.hasOwnProperty;
try {
    JSON.parse("{");
} catch (caught) {
    if (caught instanceof Error) {
        console.log(String(own.call(caught, "code")));
    }
}
