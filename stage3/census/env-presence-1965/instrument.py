import pathlib,sys,hashlib,json
p=pathlib.Path(sys.argv[1])/"src/compiler/sys.ts"
b=p.read_bytes()
needle=b"if (sys && sys.getEnvironmentVariable) {"
assert b.count(needle)==1
line=b"(globalThis as any).__envPresence1965(sys);"
nl=b"\r\n" if b"\r\n" in b else b"\n"
assert not (nl==b"\r\n" and b"\n" in b.replace(nl,b""))
a=b.replace(needle,line+nl+needle)
p.write_bytes(a)
print(json.dumps({"before":hashlib.sha256(b).hexdigest(),"after":hashlib.sha256(a).hexdigest(),"insert":line.decode(),"originalConditionRetained":True},indent=2))
