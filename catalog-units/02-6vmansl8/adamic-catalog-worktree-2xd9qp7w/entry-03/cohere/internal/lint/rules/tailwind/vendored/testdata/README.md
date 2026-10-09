# tailwindcss 4.3.3, vendored for tests

These files are copied unchanged from the published npm package tailwindcss@4.3.3
(https://www.npmjs.com/package/tailwindcss/v/4.3.3, source https://github.com/tailwindlabs/tailwindcss),
which is MIT licensed. Its LICENSE is kept beside them in tailwindcss/LICENSE.

Only what cohere's Tailwind design-system loader reads is here: index.css and the stylesheets it
imports (theme.css, preflight.css, utilities.css), plus package.json. Nothing here is built into
cohere's binary; it is test input.

To refresh to another version, copy the same six files from that version's installed package and
update TailwindVersion in vendored.go.
