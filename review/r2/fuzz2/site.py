#!/usr/bin/env python3
# site.py <tree> <program.a>...: for a program that stopped at an inserted check, find where. The
# program's JavaScript backend output is run on Node with its checks reporting their caller's line
# and column instead of panicking, and the line of program.mjs there is printed: the expression the
# check guards. Index checks report through adamicSetIndex the same way.
import os, re, subprocess, sys, tempfile
tree = sys.argv[1]
adamic = os.path.join(tree, 'adamic-cli')
for program in sys.argv[2:]:
    js = subprocess.run([adamic, 'js', program], capture_output=True, text=True).stdout
    report = ("const adamicReport = (message) => { const frame = new Error().stack.split('\\n')[3]; "
              "process.stderr.write(`CHECK ${message} AT ${frame.trim()}\\n`); process.exit(70); };\n")
    js = js.replace("const adamicDefined = (value, message) => value === undefined ? panic(message) : value;",
                    report + "const adamicDefined = (value, message) => value === undefined ? adamicReport(message) : value;")
    js = js.replace("if (!(Number.isInteger(index) && index >= 0 && index < array.length)) panic(",
                    "if (!(Number.isInteger(index) && index >= 0 && index < array.length)) adamicReport(")
    with tempfile.TemporaryDirectory(dir=os.environ.get('TMPDIR')) as work:
        path = os.path.join(work, 'program.mjs')
        open(path, 'w').write(js)
        ran = subprocess.run(['node', '--disable-warning=ExperimentalWarning', os.path.join(tree, 'oracle', 'node.mjs'), path], capture_output=True, text=True)
    match = re.search(r'CHECK (.*) AT .*program\.mjs:(\d+):(\d+)', ran.stderr)
    if not match:
        print(f'{os.path.basename(program)}\tNO SITE\t{ran.stderr.strip()[:200]}')
        continue
    line = js.split('\n')[int(match.group(2)) - 1]
    column = int(match.group(3))
    print(f'{os.path.basename(program)}\t{match.group(1)[:60]}\t{line.strip()[:220]}')
