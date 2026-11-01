# Rerun at 8a7cd267: the message without String()

The same checks as the parent directory, on the helper whose null and undefined
branch writes `value === null ? "null" : "undefined"`. apply.sh exits 0; a second
adapter run writes zero files; 21 of 21 witnesses agree; the mutant goes red on
null and undefined, 19 of 21, exit 1; the own-file table is identical to main's,
145 in 21 of 79 files, hostErrors.ts 0; stock npm ci and npm run build exit 0.
The census here is built at main 4d63c0a6, without Lower's #bbhwj30, and still
reads hostErrors.ts Refused at 14:15, "throwing a TypeError".
