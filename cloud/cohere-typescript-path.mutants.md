# Mutants for cohere's TypeScript path reader (#g7nnryf)

`cloud/cohere-typescript-path.sh` and the scripts that read through it, `stage3/meter/twice-daily.sh` among them (its
per-ref insteadOf now names `cohere/$(cohere_typescript_path)`), all lean on the helper's one answer. Two mutants show
its test can fail.

## The entry chosen by its url, not its name

`cohere_typescript_path`'s loop, old (this head):

```sh
	while IFS=$'\n' read -r -d '' key value; do
		name=${key#submodule.}
		name=${name%.path}
		...
		if [ "$name" = TypeScript ]; then
			paths+=("$value")
		fi
	done < <(git config --no-includes --file "$gitmodules" -z --get-regexp '^submodule\..*\.path$' 2> /dev/null)
```

New, 296c1844's reader put back: the entry whose url's last part is TypeScript, whatever its name.

Run: `bash cloud/cohere-typescript-path-test.sh`, sealed, at the mutant's commit 67a6b1cf88ecd17ef7e99e0272ea91a14f1ab260
(Loom job J-v2q1bs1).

Red, exit 1, in 4 cases. The renamed one:
`red  renamed: wanted FAIL holding 'has 0 submodule paths named TypeScript, ...', got exit 0, stdout 'libraries/typescript', stderr ''`
The others are none, two and lowercase. The run ends `FAIL: the helper went red in 4 cases`.

## A hard-coded path

The test's own mutant, run in the same file each time: the helper replaced by one that prints `TypeScript` whatever
`.gitmodules` says. It goes red in 12 cases, the moved layouts at `libraries/typescript` and
`libraries/typescript/TypeScript` among them.

Green at this head: Loom job J-y5b306f at fe811e2630c30397c3d7739f91da506ffabb2528 passes every case, its log ending
`the helper passes every case; the mutant goes red in 12, the moved shape among them`, then `PASS`.
