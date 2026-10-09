#!/bin/zsh
# cohere against the toolchains it replaces, timed the same way on the same pinned copy of a project.
#
# Three stacks, each asked the same question, read-only: does this tree lint clean, type-check and sit
# formatted? Nothing is fixed or written.
#   cohere   `cohere --no-fix --format`: lint, types and the format check in one process
#   oxc      `oxlint --type-aware --type-check` (type-aware rules through tsgolint, and the compiler's
#            own diagnostics), then `oxfmt --check`
#   eslint   `eslint` with the project's own config (typescript-eslint and its type-aware rules), then
#            `prettier --check .`, then `tsc --noEmit`
# A stack's time is its tools' times added up, because that is how a gate runs them, one after another.
# Every tool's own time is printed too.
#
# Every rival is the stock release at a pinned version, installed into a tools directory of its own and
# run from there, and configured the way someone adopting it on this project would. Nothing is installed
# into the project, and none of the project's own copies are used: ahra's `prettier` is a house fork
# (`@system-inc/prettier`), and its ESLint config no longer loads under ESLint, so neither is what a
# reader comparing against "ESLint and Prettier" would have. The configurations differ, and the
# numbers have to be read with that:
#   - ESLint runs typescript-eslint's `recommendedTypeChecked`, typed through the project service, over
#     the files the project's tsconfig covers (its excludes, as ignores). These are not the project's
#     rules.
#   - oxlint runs its default rules plus every type-aware rule those defaults enable. Not the project's
#     rules either, and neither rival's findings are comparable to cohere's.
#   - Both formatters get the options cohere formats with, the format block of cohere's TypeScript set
#     read at the engine's commit, so all three format to the same width, indent, quotes and semicolons.
#     Prettier also gets its Tailwind plugin, pointed at the project's own Tailwind entry point and class
#     functions; oxfmt gets no plugin. Both check exactly the files the project's tsconfig puts in the
#     program, less the format block's ignores, named one by one, so neither formats a file cohere would
#     not. tsc gets the project's tsconfig.
#   - A project whose house format is not stock Prettier's fails both format checks on many files. That
#     does not shorten the run: a check formats every file to compare it. Exit codes are printed so a
#     failing check is never mistaken for a passing one.
#
# Cold and warm. Cold gives every tool no cache: cohere `--no-cache`, ESLint and Prettier without
# `--cache`, tsc with `--incremental false`. Warm gives every tool the cache it offers, primed by an
# unmeasured run just before: cohere's own, ESLint's and Prettier's `--cache` at a location under the
# logs, tsc's incremental build info under the logs. oxlint and oxfmt keep no cache, so their warm run is
# a second run with the files already in memory, and the table says `none`.
#
# Every time here is a run to its end, cache write included, because that is the only end the other
# tools have. cohere's launcher answers before its cache write, so a developer waits less than cohere's
# column says; quiet_machine.sh measures that.
#
# Stacks rotate their order every round, and every run is printed with the one-minute load before and
# after it, labelled quiet or loaded against the ceiling exactly as quiet_machine.sh does.
#
# usage: rivals.sh [options]
#   --project DIR      the project to copy (default ~/Projects/ahra)
#   --runs N           rounds (default 3)
#   --load-ceiling L   the highest one-minute load a quiet run may start or end at (default: cores / 4)
#   --settle SECONDS   before each measured run, wait up to this long for the load to reach the ceiling
#   --cohere PATH      the cohere launcher, or an engine binary (default: `cohere` on PATH)
#   --work DIR         where the copies live (default $TMPDIR/cohere-quiet-machine, shared with
#                      quiet_machine.sh so both measure the same copy)
#   --tools DIR        where the rivals are installed (default <work>/rival-tools)
#   --only TOOLS       measure only these tools, comma-separated (cohere, oxlint, oxfmt, eslint,
#                      prettier, tsc); a stack left with none is skipped
# The versions installed are the defaults below; edit them here, since every one is part of the result.
#
# Exit status: 0 when every stack and mode has a quiet number, 3 when one has none, 1 when it could not
# measure at all.
set -u
name=rivals
source ${0:A:h}/pinned.zsh

project=$HOME/Projects/ahra
runs=3
ceiling=$(awk -v cores=$(sysctl -n hw.ncpu) 'BEGIN { print cores / 4 }')
settle=0
cohere=cohere
work=${TMPDIR:-/tmp}/cohere-quiet-machine
tools=
only=(cohere oxlint oxfmt eslint prettier tsc)
rivals=(
  oxlint@1.86.0 oxlint-tsgolint@7.0.2003 oxfmt@0.71.0
  eslint@10.8.1 typescript-eslint@8.67.0 typescript@6.0.3 prettier@3.9.6 prettier-plugin-tailwindcss@0.8.1
)

while (( $# > 0 )); do
  case $1 in
    --project) project=$2; shift 2 ;;
    --runs) runs=$2; shift 2 ;;
    --load-ceiling) ceiling=$2; shift 2 ;;
    --settle) settle=$2; shift 2 ;;
    --cohere) cohere=$2; shift 2 ;;
    --work) work=$2; shift 2 ;;
    --tools) tools=$2; shift 2 ;;
    --only) only=(${(s:,:)2}); shift 2 ;;
    *) fail "unknown argument $1 (see the usage at the top of this script)" ;;
  esac
done
(( runs >= 1 )) || fail "--runs must be at least 1"

prepare_copy
logs=$copy/.cache/rivals-$(date +%Y%m%d-%H%M%S)
mkdir -p $logs
runs_table=$logs/runs.tsv
pin_engine $logs

# The tools directory is keyed by the versions, so changing one installs a fresh set rather than mixing
# versions in one node_modules.
tools=${tools:-$work/rival-tools}/$(print -l $rivals | shasum -a 256 | cut -c1-12)
if [[ ! -f $tools/installed ]]; then
  print -u2 "installing ${(j:, :)rivals} into $tools"
  mkdir -p $tools
  print '{ "private": true }' > $tools/package.json
  (cd $tools && npm install --no-audit --no-fund --no-save $rivals > install.log 2>&1) ||
    fail "installing the rivals failed, see $tools/install.log"
  print -l $rivals > $tools/installed
fi
bin=$tools/node_modules/.bin
# oxlint looks for tsgolint beside the project it lints, which is the one place it is not, so it is named
# outright: the native binary, not the node wrapper in front of it.
tsgolint_binary=($tools/node_modules/@oxlint-tsgolint/*/tsgolint(N))
(( ${#tsgolint_binary} == 1 )) || fail "no native tsgolint binary under $tools"

# Each run writes the configurations it gives the rivals, from the project's own settings, beside its
# logs, except ESLint's, which has to sit in the tools directory to import typescript-eslint from it.
# Paths in them are absolute, because each is read from outside the project it describes.
#
# The format options are cohere's own, from the cohere checkout the engine was built from, at the commit
# it reports, so the rivals format exactly as cohere does on this run. The Tailwind plugin is resolved in
# the tools directory, since by name it would find the project's own, built against another Prettier.
cohere_checkout=${engine_source:h:h:h:h}
cohere_commit=$(print -- "$cohere_version" | awk '$1 == "commit:" { print $2 }')
format_set=$(git -C $cohere_checkout show $cohere_commit:internal/lint/configuration/sets/typescript.json) ||
  fail "cannot read cohere's TypeScript set at $cohere_commit in $cohere_checkout"
(cd $copy && print -r -- "$format_set" | node -e '
  const [copy, logs, tools] = process.argv.slice(1);
  const fs = require("fs");
  const format = JSON.parse(fs.readFileSync(0, "utf8")).format;
  if (!format) throw new Error("cohere'\''s TypeScript set has no format block");
  const { ignore = [], ...options } = format;
  fs.writeFileSync(logs + "/oxfmtrc.json", JSON.stringify(options));
  fs.writeFileSync(logs + "/format-ignore.txt", ignore.join("\n") + "\n");
  const tailwind = JSON.parse(fs.readFileSync("CohereSettings.json", "utf8")).settings?.["better-tailwindcss"] ?? {};
  const prettier = { ...options, plugins: [require.resolve("prettier-plugin-tailwindcss", { paths: [tools] })] };
  if (tailwind.entryPoint) prettier.tailwindStylesheet = copy + "/" + tailwind.entryPoint.replace(/^\.\//, "");
  if (tailwind.callees) prettier.tailwindFunctions = tailwind.callees;
  fs.writeFileSync(logs + "/prettierrc.json", JSON.stringify(prettier));
' $copy $logs $tools) || fail "could not write the formatters' options"
# The files: the program the project's tsconfig defines, its own files only, less the format ignores.
(cd $copy && $bin/tsc -p tsconfig.json --listFilesOnly) | awk -v root=$copy/ 'index($0, root) == 1 && $0 !~ /\/node_modules\// { print substr($0, length(root) + 1) }' |
  while read -r file; do
    skip=
    for pattern in ${(f)"$(< $logs/format-ignore.txt)"}; do [[ ${file:t} == ${~pattern} ]] && skip=1; done
    [[ -z $skip ]] && print -r -- $file
  done > $logs/format-files.txt
format_file_count=$(wc -l < $logs/format-files.txt | tr -d ' ')
(( format_file_count > 0 )) || fail "the project's tsconfig names no files to format"
# The ignores are the project tsconfig's excludes, so ESLint lints the files the type checker covers.
cat > $tools/eslint.config.mjs <<EOF
import { defineConfig } from 'eslint/config';
import typeScriptEslint from 'typescript-eslint';
export default defineConfig(
  {
    basePath: '$copy',
    ignores: ['public/**', '.next/**', '.open-next/**', '.vercel/**', '.wrangler/**', '.cache/**', 'projects/**',
      'data/**', 'modules/*/data/**', 'modules/*/*/data/**'],
  },
  {
    basePath: '$copy',
    files: ['**/*.ts', '**/*.tsx', '**/*.mts', '**/*.mjs'],
    extends: [typeScriptEslint.configs.recommendedTypeChecked],
    languageOptions: { parserOptions: { projectService: true, tsconfigRootDir: '$copy' } },
  },
);
EOF

print "# rivals benchmark, $(date -u +%Y-%m-%dT%H:%M:%SZ)"
print_pinned_header
print "# rivals   ${(j:, :)rivals}"
print "#          from $tools"
print "#          format options $(< $logs/oxfmtrc.json), from cohere's TypeScript set at $cohere_commit"
print "#          prettier $(< $logs/prettierrc.json)"
print "#          formatters check $format_file_count files, the tsconfig's program less the format ignores"
print "#          measuring ${(j:, :)only}"
print "#          eslint: typescript-eslint recommendedTypeChecked, $tools/eslint.config.mjs"
print "# rounds   $runs, load ceiling $ceiling (one-minute)"
print

# tool_commands prints a stack's tools for one mode, one per line: a label, a tab, then the command.
tool_commands() {
  local stack=$1 mode=$2
  case $stack-$mode in
    cohere-cold) print "cohere\t$engine --no-fix --format --no-cache" ;;
    cohere-warm) print "cohere\t$engine --no-fix --format" ;;
    oxc-*)
      print "oxlint\tenv OXLINT_TSGOLINT_PATH=$tsgolint_binary[1] $bin/oxlint --type-aware --type-check"
      print "oxfmt\t$bin/oxfmt --check -c $logs/oxfmtrc.json @FILES" ;;
    eslint-cold)
      print "eslint\t$bin/eslint --config $tools/eslint.config.mjs ."
      print "prettier\t$bin/prettier --check --config $logs/prettierrc.json @FILES"
      print "tsc\t$bin/tsc --noEmit --incremental false" ;;
    eslint-warm)
      print "eslint\t$bin/eslint --config $tools/eslint.config.mjs . --cache --cache-location $logs/eslint-cache/"
      print "prettier\t$bin/prettier --check --config $logs/prettierrc.json --cache --cache-location $logs/prettier-cache @FILES"
      print "tsc\t$bin/tsc --noEmit --incremental --tsBuildInfoFile $logs/tsc.tsbuildinfo" ;;
  esac | while IFS=$'\t' read -r tool command; do
    (( ${only[(Ie)$tool]} )) && print -r -- "$tool"$'\t'"$command"
  done
}

# run_stack runs each tool of a stack and sets stack_seconds to their sum and stack_exits to their exit
# codes. With a label it prints one row per tool; without, the run is a prime and prints nothing.
run_stack() {
  local stack=$1 mode=$2 round=$3 label=${4:-}
  stack_seconds=0 stack_exits=()
  tool_commands $stack $mode | while IFS=$'\t' read -r tool command; do
    local log=$logs/$stack-$mode-$round-$tool.log
    [[ -z $label ]] && log=$logs/$stack-$mode-$round-$tool.prime.log
    # `@FILES` stands for the formatters' file list, one argument per file.
    local -a words=(${(z)command})
    local at=${words[(Ie)@FILES]}
    (( at )) && words[$at,$at]=(${(f)"$(< $logs/format-files.txt)"})
    run_timed $log $words
    stack_seconds=$(( stack_seconds + settled_seconds ))
    stack_exits+=($exit_code)
    [[ -n $label ]] && printf '%s\t%s\t%d\t%s\t%.3f\t%d\n' $stack $mode $round $tool $settled_seconds $exit_code >> $logs/tools.tsv
  done
}

measure() {
  local stack=$1 mode=$2 round=$3
  [[ $mode == warm ]] && run_stack $stack $mode $round
  wait_for_quiet
  local before=$(load)
  run_stack $stack $mode $round measured
  local after=$(load)
  local quiet=loaded
  under_ceiling $before && under_ceiling $after && quiet=quiet
  printf '%s\t%s\t%d\t%.3f\t%s\t%s\t%s\t%s\n' $stack $mode $round $stack_seconds "${(j:,:)stack_exits}" \
    $before $after $quiet | tee -a $runs_table
}

stacks=(cohere oxc eslint)
printf 'stack\tmode\tround\tseconds\texits\tload_before\tload_after\tquiet\n'
for round in $(seq 1 $runs); do
  # Rotated, so no stack always runs first, or always runs after the heaviest.
  order=(${stacks[$(( (round - 1) % 3 + 1 )),-1]} ${stacks[1,$(( (round - 1) % 3 ))]})
  for mode in cold warm; do
    for stack in $order; do
      [[ -n $(tool_commands $stack $mode) ]] && measure $stack $mode $round
    done
  done
done

[[ $(tree_hash) == $expected_tree ]] ||
  fail "the copy's files changed during the benchmark, so the runs did not all see the same tree; see $logs"

print
print "each tool's seconds, every run:"
awk -F'\t' '{ key = $1 " " $2 " " $4; times[key] = times[key] (times[key] == "" ? "" : ", ") $5; exits[key] = exits[key] (exits[key] == "" ? "" : ",") $6 }
  END { for (key in times) printf "  %-22s %s (exit %s)\n", key, times[key], exits[key] }' $logs/tools.tsv | sort

print
print "stack seconds (quiet runs are the number; loaded runs are shown, not reported):"
outcome=0
for stack in $stacks; do
  for mode in cold warm; do
    quiet_runs=($(awk -F'\t' -v stack=$stack -v mode=$mode '$1 == stack && $2 == mode && $8 == "quiet" { print $4 }' $runs_table | sort -n))
    loaded_runs=($(awk -F'\t' -v stack=$stack -v mode=$mode '$1 == stack && $2 == mode && $8 == "loaded" { print $4 }' $runs_table | sort -n))
    (( ${#quiet_runs} + ${#loaded_runs} > 0 )) || continue
    if (( ${#quiet_runs} > 0 )); then
      printf '  %-6s %-4s quiet best %s, median %s, worst %s (%d of %d runs quiet)\n' $stack $mode \
        $quiet_runs[1] $quiet_runs[$(( (${#quiet_runs} + 1) / 2 ))] $quiet_runs[-1] ${#quiet_runs} $runs
    else
      outcome=3
      printf '  %-6s %-4s no quiet number: all %d runs started or ended above the load ceiling of %s\n' $stack $mode $runs $ceiling
    fi
    (( ${#loaded_runs} > 0 )) && printf '              loaded runs, not a quiet number: %s\n' "${(j:, :)loaded_runs}"
  done
done
print "runs and logs: $logs"
exit $outcome
