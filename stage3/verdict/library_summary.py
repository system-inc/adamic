"""Reproduce upstream default-library summary placeholders and virtual ordering."""
import re

HEADER = re.compile(rb'^([^\n]+?)\((\d+),(\d+)\): (?:error|message) TS(\d+):', re.M | re.I)


def project(stdout, library_names):
    if not library_names:
        return stdout
    records = []
    matches = list(HEADER.finditer(stdout))
    if not matches:
        return stdout
    # Never discard output outside diagnostic blocks.
    if matches[0].start() != 0:
        return stdout
    for index, match in enumerate(matches):
        end = matches[index + 1].start() if index + 1 < len(matches) else len(stdout)
        block = stdout[match.start():end]
        filename = match[1]
        basename = filename.rsplit(b'/', 1)[-1].decode(errors='replace')
        if basename in library_names:
            location = basename.encode() + b'(--,--):'
            block = location + block[match.end(3) + 2 - match.start():]
            filename = b'/.ts/' + basename.encode()
        elif not filename.startswith(b'/'):
            filename = b'/.src/' + filename
        records.append(((filename, int(match[2]), int(match[3]), int(match[4])), block))
    return b''.join(block for _, block in sorted(records, key=lambda record: record[0]))
