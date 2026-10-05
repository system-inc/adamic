#!/usr/bin/env python3
# interact.py <command...>: start a program with stdin and stdout pipes, wait up to five seconds for its
# first line, and only then answer it. A program that holds its prompt in a buffer never shows it.
import select, subprocess, sys
process = subprocess.Popen(sys.argv[1:], stdin=subprocess.PIPE, stdout=subprocess.PIPE)
ready, _, _ = select.select([process.stdout], [], [], 5)
if not ready:
    print("no prompt within 5s; answering anyway")
else:
    print("prompt:", process.stdout.readline().decode().rstrip())
process.stdin.write(b"yes\n")
process.stdin.close()
print("rest:", process.stdout.read().decode().rstrip(), "exit", process.wait())
