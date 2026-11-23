const parsed = JSON.parse('{"message":"x","hasOwnProperty":5}') as Error;
try {
    console.log(String(parsed.hasOwnProperty("message")));
} catch (caught) {
    console.log("threw");
}
