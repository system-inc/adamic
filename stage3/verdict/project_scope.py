"""Append an explicit virtual file list without moving original JSONC locations."""
import json


def append_files(text, names):
    depth = 0
    quote = escaped = line_comment = block_comment = False
    previous = ''
    index = 0
    while index < len(text):
        char = text[index]
        following = text[index:index + 2]
        if line_comment:
            line_comment = char not in '\r\n'
        elif block_comment:
            if following == '*/':
                block_comment = False
                index += 1
        elif quote:
            if escaped:
                escaped = False
            elif char == '\\':
                escaped = True
            elif char == '"':
                quote = False
        elif following == '//':
            line_comment = True
            index += 1
        elif following == '/*':
            block_comment = True
            index += 1
        elif char == '"':
            quote = True
            previous = char
        elif char == '{':
            depth += 1
            previous = char
        elif char == '}':
            depth -= 1
            if depth == 0:
                comma = '' if previous in ('{', ',') else ','
                return text[:index] + comma + '\n"files": ' + json.dumps(names) + '\n' + text[index:]
            previous = char
        elif not char.isspace():
            previous = char
        index += 1
    raise RuntimeError('namespace project lacks closing object')
