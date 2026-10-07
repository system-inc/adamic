export const messageUnicodeBomExpected =
    'This file is configured to start with a byte order mark and does not. The mark is what tells a reader that has no other signal which encoding and endianness the bytes are in, so a project that requires it wants every file to carry it rather than most.';
export const messageUnicodeBomUnexpected =
    'This file starts with a byte order mark. In a project that is already UTF-8 everywhere the mark carries no information and is invisible in an editor, while tools that read the first bytes literally see it: a shell script stops being executable, a concatenated bundle gains a stray character mid-file, and a diff shows a change on a line nobody edited.';
