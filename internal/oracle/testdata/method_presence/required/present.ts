interface Host { read(): string }
function probe(host: Host): string {
    return host.read ? host.read() : "absent";
}
console.log(probe({ read() { return "present"; } }));
