function test(): void {
    const mappingCharCodes: number[] = [65, 66];
    let mappings = "";
    function flushMappingBuffer(): void {
        if (mappingCharCodes.length > 0) {
            mappings += String.fromCharCode.apply(undefined, mappingCharCodes);
            mappingCharCodes.length = 0;
        }
    }
    flushMappingBuffer();
    console.log(mappings);
}
test();
