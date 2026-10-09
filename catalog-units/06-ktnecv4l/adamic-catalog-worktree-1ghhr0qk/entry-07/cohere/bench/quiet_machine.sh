#!/bin/zsh
# The quiet-machine benchmark: how long `cohere` takes on a real project, cold, on an unchanged replay,
# and after a one-file edit, measured so that the number means what it says.
#
# It runs at the priority it is started with, never niced as the house's test and build runs are
# (cohere-dev test, #2qc6j8g): it measures, and a niced run's times say how busy the machine was rather
# than how fast cohere is. Run it only in a quiet window, where nothing else is competing.
#
# It measures a copy of the project pinned to its commits, never the project itself, with one cohere
# engine for every run. pinned.zsh, which it shares with rivals.sh, says how and why.
#
# The command measured is `cohere --json`: the one a developer runs, so a warm run gets the caches a
# developer's run would, printing its versioned contract instead of the footer. The footer is for people and
# free to change, and when it became one line, everything this benchmark read from it read nothing
# (#zrgrk14), so each run's engine seconds, findings and cache use are read from its --json summary by
# internal/benchresults/tools/runoutput, which is tested against real output. That command may write, so
# the copy's files are hashed before the first run and after the last one, and a benchmark that changed a
# byte of them refuses to report.
#
# One round, interleaved, so a load that drifts lands on every mode alike. A round measures only the modes
# asked for with --modes, cold alone unless told otherwise: cold is the number this benchmark exists for, and
# every other mode costs wall time a quiet window has to hold for (#bdqqfa1).
#   prime   `cohere`, not measured, so the cache describes the tree as it stands: run again until one
#           replays, up to three times. Only a round that measures warm or edit primes, since only those
#           read the cache, and the prime is the check that the cache is whole: the loop stops at the first
#           one that replays, so a cache already whole costs one replay, and nothing is skipped on trust. A
#           second prime is expected every round after the first: the edit run's cache entry holds the
#           edited bytes, so the next run rechecks the restored file and its importers. Each run's row carries
#           how many primes its round needed (0 when the round reads no cache), every prime's cache use is
#           kept in primes.tsv, and the summary warns on what is actually wrong: a prime after the first round
#           that took nothing from the cache (#547dhjz), or a warm run that is not a whole replay
#   warm    `cohere` again on the unchanged tree: the replay
#   edit    one line added to one function body, `cohere`, then the file's original bytes put back
#   cold    `cohere --no-cache`: every phase from source, nothing read or written
#   idle    `cohere --no-cache` after the machine has sat idle for --idle-seconds: cold as a developer who
#           pauses between saves mostly meets it, since the slow mode below follows idleness
#   first   `cohere` in a fresh copy of the tree with no cache of any kind: what a developer's first run
#           pays, cache writes and incremental build info included. Measured beside cold, because a first
#           run doing more than cold is a cliff nothing else here would show: on 2026-10-04 a first run
#           on www-ahra-ai took 8.6s against 0.8s cold (#0q6nmnt). Its runs go to first.tsv, apart from
#           the runs table the record reads, and the summary flags a quiet median more than 25% over
#           cold's. Its template, a cache-free clone of the copy, is made and checked only when first is asked for
# The edit is a comment line, inserted above a `return` statement, so the file's bytes and every
# content hash cohere keys on change while its types and findings do not. A round whose edit run
# reports a different finding count from its warm run says so, because then the edit did more than that.
#
# A run is quiet when the machine was: the share of all cores idle, sampled over the 2 seconds just before
# it and the second just after it, at or over the floor (--idle-floor). Before each measured run the benchmark
# samples until the floor is met, for up to --settle seconds, then runs either way and labels the run by
# what it measured. A best or median is only called quiet when it is taken from quiet runs, and a mode with
# no quiet runs gets no quiet number, only the loaded ones labelled as loaded, because a best-of-3 taken on a
# busy machine and one taken on a quiet one look the same afterwards. The sample after the run is what
# catches a build that started while it ran.
#
# Idleness replaced the one-minute load average, which every run used to wait on (record schema 1). That
# average mostly measures cohere's own last run: a two-second cold run on every core lifts it by about one,
# and it decays for 30 to 60 seconds, so each run waited up to the settle for the benchmark's own echo. On
# 2026-10-06 five rounds of every mode took about three and a half minutes, most of it waiting. The load is
# still printed and recorded beside each run, and decides nothing.
#
# Rounds stop early once every measured mode has at least 3 quiet runs within --band percent of their
# median, best to worst. The record says it stopped early, and validating it checks that the runs agree.
#
# Two times per run. `verdict` is what a developer waits for: the launcher returns when the engine has
# its exit code, and the engine writes its cache afterwards. `settled` is when every process the run
# started has exited, cache write included. The next run starts only after that. The engine runs
# directly with the verdict descriptor the launcher would have given it, so the early return is still
# what is timed.
#
# Cold runs are bimodal on this Mac, and the slow mode follows idleness, so back-to-back runs understate it
# (#g3046x5, 2026-10-06). In about half of quiet cold runs on ahra every file syscall costs about three times
# as much from the process's first milliseconds: discovery's sys 0.5 to 0.95s against 0.2 to 0.3s, the
# graph's 2.1 to 2.9s against 1.2 to 1.4s, about 0.1s of wall. It is set before the build and moved by
# nothing in it: capping the build's concurrent reads at 12 or 8, or turning the early format pass off, left
# its rate where it was. What moves it is how long the machine sat idle before the run. Chained cold runs at
# 16 threads, gaps rotated, took it 2 of 6 times one second after the previous run exited, 2 of 5 at ten
# seconds, and 5 of 5 at sixty. It also needs every core: at two threads the same discovery costs 0.05s of
# sys, and under heavy load the mode never shows. So cold, whose runs follow one another within seconds,
# samples the fast mode more often than a developer who saves and checks after half a minute of thought, and
# its median is the better case. The idle mode measures the other one.
#
# usage: quiet_machine.sh [options]
#   --project DIR        the project to copy (default ~/Projects/ahra)
#   --modes LIST         the modes to measure, comma-separated: cold, warm, edit, idle, first (default cold)
#   --runs N             at most N rounds, so at most N runs of each mode (default 5)
#   --band PERCENT       stop once every measured mode's quiet runs, at least 3 of them, are within this
#                        percent of their median, best to worst (default 10; 0 never stops early)
#   --idle-floor PERCENT the share of all cores idle before and after a quiet run (default 90)
#   --settle SECONDS     before each measured run, sample idleness for up to this long until it reaches the
#                        floor, then run either way (default 60)
#   --idle-seconds S     how long the machine sits idle before an idle run (default 60)
#   --edit PATH          the file the edit run changes, relative to the project
#                        (default modules/tasks/TasksSearchQuery.ts)
#   --cohere PATH        the cohere launcher, or an engine binary, to run (default: `cohere` on PATH)
#   --work DIR           where the copies live (default $TMPDIR/cohere-quiet-machine)
#   --record             also write the run to bench/results as a record to commit
#                        (internal/benchresults, schema 2): the machine, the commits, the modes, every run
#                        and each mode's numbers, refused when it disagrees with its own runs or the engine
#                        is dirty; commit it with what `go run ./internal/docsdata/tools/generate` rewrites
#
# Exit status: 0 when every measured mode has a quiet number, 3 when some mode has none, 4 when every mode has
# one and the first run's quiet median is more than 25% over cold's, 1 when it could not measure at all,
# including a copy that changed under it or a cohere that changed between runs.
set -u
name=quiet_machine
source ${0:A:h}/pinned.zsh

project=$HOME/Projects/ahra
modes_asked=cold
runs=5
band=10
idle_floor=90
settle=60
idle_seconds=60
edit=modules/tasks/TasksSearchQuery.ts
cohere=cohere
work=${TMPDIR:-/tmp}/cohere-quiet-machine
record=false

while (( $# > 0 )); do
  case $1 in
    --project) project=$2; shift 2 ;;
    --modes) modes_asked=$2; shift 2 ;;
    --runs) runs=$2; shift 2 ;;
    --band) band=$2; shift 2 ;;
    --idle-floor) idle_floor=$2; shift 2 ;;
    --settle) settle=$2; shift 2 ;;
    --idle-seconds) idle_seconds=$2; shift 2 ;;
    --edit) edit=$2; shift 2 ;;
    --cohere) cohere=$2; shift 2 ;;
    --work) work=$2; shift 2 ;;
    --record) record=true; shift ;;
    *) print -u2 "quiet_machine: unknown argument $1 (see the usage at the top of this script)"; exit 1 ;;
  esac
done
(( runs >= 1 )) || fail "--runs must be at least 1"
modes=(${(s:,:)modes_asked})
(( ${#modes} > 0 )) || fail "--modes names no mode"
for mode in $modes; do
  [[ $mode == (cold|warm|edit|idle|first) ]] || fail "--modes: no mode $mode (cold, warm, edit, idle, first)"
done
measuring() { (( ${modes[(Ie)$1]} )); }
# The modes a record holds: the first runs have a table of their own and are never recorded.
record_modes=(${modes:#first})
[[ $record == false ]] || (( ${#record_modes} > 0 )) || fail "--record needs a mode other than first"

prepare_copy
measuring edit && { [[ -f $copy/$edit ]] || fail "no file $edit in the copy to edit"; }

# The first mode's template: the copy without its cache, made once per copy and cloned for each first run,
# so a first run never sees what an earlier run of any mode left behind. Its tree is the copy's, or it is
# refused, the same as the copy. Made and checked only when first is measured: the copy and two tree hashes
# cost seconds every other run would pay for nothing.
#
# Its ready marker is written here, never copied: written only once the template's tree matches the copy's,
# so a copy that died part way is never taken for a finished one. And a TypeScript build info anywhere in it
# (tsc's default puts one beside the tsconfig, outside .cache) would make a first run not first, so a
# template holding one is refused rather than measured.
if measuring first; then
  template=$work/${copy:t}-first-template
  if [[ ! -f $template/bench-ready ]]; then
    [[ -e $template ]] && fail "$template exists but was never finished; remove it and run again"
    mkdir -p $template
    rsync -a --exclude=/.cache --exclude=/bench-ready "$copy/" "$template/" || fail "copying the copy into $template"
    [[ $(copy=$template tree_hash) == $expected_tree ]] ||
      fail "the first runs' template at $template does not match the copy; remove it and run again"
    print $expected_tree > $template/bench-ready
  fi
  [[ $(copy=$template tree_hash) == $(< $template/bench-ready) && $(< $template/bench-ready) == $expected_tree ]] ||
    fail "the first runs' template at $template no longer matches the copy; remove it and run again"
  build_infos=(${(f)"$(cd $template && find . -name '*.tsbuildinfo' -not -path './.cache/*' -not -path '*/node_modules/*')"})
  (( ${#build_infos} == 0 )) ||
    fail "the copy holds a TypeScript build info outside .cache (${build_infos[1]}), so a first run there would not be first"
fi

logs=$copy/.cache/quiet-machine-$(date +%Y%m%d-%H%M%S)
mkdir -p $logs
runs_table=$logs/runs.tsv
pin_engine $logs
first_table=$logs/first.tsv
run_output_reader=$logs/runoutput
(cd ${0:A:h:h} && go build -o $run_output_reader ./internal/benchresults/tools/runoutput) ||
  fail "could not build the run output reader"

print "# quiet-machine benchmark, $(date -u +%Y-%m-%dT%H:%M:%SZ)"
print_pinned_header
print "# modes    ${(j:, :)modes}; at most $runs rounds, stopping once quiet runs agree within $band%"
print "# quiet    at least $idle_floor% of all cores idle over 2s before and 1s after a run, settling up to ${settle}s"
measuring edit && print "# edit     $edit"
measuring idle && print "# idle     each idle run after ${idle_seconds}s of the machine sitting idle"
print

# run_cohere runs the pinned engine in the copy; see run_timed.
run_cohere() {
  local log=$1; shift
  run_timed $log $engine --json "$@"
}

at_floor() { awk -v idle=$1 -v floor=$idle_floor 'BEGIN { exit !(idle >= floor) }'; }

# settle_until_idle samples idleness over 2 seconds until it reaches the floor, for up to `settle` seconds,
# and sets idle_before to the last sample: the run that follows is quiet only if that sample was.
settle_until_idle() {
  local deadline=$(( EPOCHREALTIME + settle ))
  idle_before=$(idle_over 2)
  while ! at_floor $idle_before && (( EPOCHREALTIME < deadline )); do idle_before=$(idle_over 2); done
}

measure() {
  local mode=$1 round=$2; shift 2
  local table=${measure_table:-$runs_table}
  settle_until_idle
  local before=$(load)
  local log=$logs/$mode-$round.log
  run_cohere $log "$@"
  local after=$(load) idle_after=$(idle_over 1)
  local quiet=loaded
  at_floor $idle_before && at_floor $idle_after && quiet=quiet
  # The run's engine seconds, findings and cache use, from its --json summary. A log the reader cannot
  # read stops the benchmark rather than recording a column it could not fill.
  local output engine_seconds findings cache
  output=$($run_output_reader $log) || fail "could not read the run's output, see $log"
  IFS=$'\t' read -r engine_seconds findings cache <<< $output
  printf '%s\t%d\t%.3f\t%.3f\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d\t%d\n' $mode $round $verdict_seconds $settled_seconds \
    $engine_seconds $findings "$cache" $before $after $idle_before $idle_after $quiet $exit_code $primes |
    tee -a $table
}

# quiet_runs prints a mode's quiet verdict seconds from <table>, sorted.
quiet_runs() { awk -F'\t' -v mode=$1 '$1 == mode && $12 == "quiet" { print $3 }' $2 | sort -n; }
table_of() { [[ $1 == first ]] && print $first_table || print $runs_table; }

# agreed is true once every measured mode has at least 3 quiet runs within the band, best to worst, of their
# median (the lower middle of an even count, as the summary and the record take it).
agreed() {
  (( band > 0 )) || return 1
  local mode
  for mode in $modes; do
    local quiet=($(quiet_runs $mode $(table_of $mode)))
    (( ${#quiet} >= 3 )) || return 1
    awk -v best=$quiet[1] -v worst=$quiet[-1] -v median=$quiet[$(( (${#quiet} + 1) / 2 ))] -v band=$band \
      'BEGIN { exit !(worst - best <= median * band / 100) }' || return 1
  done
}

if measuring edit; then
  edit_original=$logs/edit-original
  cp $copy/$edit $edit_original
  original_hash=$(shasum -a 256 < $edit_original)
fi

printf 'mode\tround\tverdict_s\tsettled_s\tengine_s\tfindings\tcache\tload_before\tload_after\tidle_before\tidle_after\tquiet\texit\tprimes\n'
rounds_taken=0
stopped_early=false
for round in $(seq 1 $runs); do
  rounds_taken=$round
  primes=0
  if measuring warm || measuring edit; then
    for prime in 1 2 3; do
      primes=$prime
      run_cohere $logs/prime-$round-$prime.log
      prime_cache=$($run_output_reader $logs/prime-$round-$prime.log | cut -f3) ||
        fail "could not read the prime's output, see $logs/prime-$round-$prime.log"
      printf '%d\t%d\t%s\n' $round $prime "$prime_cache" >> $logs/primes.tsv
      [[ $prime_cache == replay ]] && break
    done
  fi

  measuring warm && measure warm $round

  if measuring edit; then
    # The comment goes above the first `return` that opens a statement of its own, so it lands in a
    # function body and cannot become the body of an unbraced `if`.
    awk -v marker="cohere quiet-machine edit, round $round" '
      !done && /^[ \t]+return[ ;(]/ && previous ~ /[{;}][ \t]*$/ {
        match($0, /^[ \t]+/); print substr($0, 1, RLENGTH) "// " marker; done = 1
      }
      { print; previous = $0 }
      END { if (!done) exit 1 }' $edit_original > $logs/edit-$round ||
      fail "found no return statement in $edit to put the edit above"
    cat $logs/edit-$round > $copy/$edit
    measure edit $round
    cat $edit_original > $copy/$edit
    [[ $(shasum -a 256 < $copy/$edit) == $original_hash ]] || fail "$edit did not go back to its original bytes"
  fi

  measuring cold && measure cold $round --no-cache

  if measuring idle; then
    sleep $idle_seconds
    measure idle $round --no-cache
  fi

  if measuring first; then
    # A fresh clone of the template for each first run (copy-on-write where the volume allows it), measured
    # there, then held to the copy's tree: a first run must write nothing but its cache.
    first_copy=$logs/first-$round
    cp -c -R $template $first_copy 2> /dev/null || cp -R $template $first_copy || fail "cloning $template"
    # A first run has no primes, so its row says 0. The clone stays in the logs with its cache, for a slow
    # first run to be read; on a volume without clones that is a full tree per round, under the logs.
    copy=$first_copy measure_table=$first_table primes=0 measure first $round
    [[ $(copy=$first_copy tree_hash) == $expected_tree ]] ||
      fail "the first run in round $round changed the tree it ran in; see $first_copy"
  fi

  if (( round < runs )) && agreed; then
    stopped_early=true
    print "stopped after round $round of $runs: every measured mode has 3 or more quiet runs within $band% of their median"
    break
  fi
done

[[ $(tree_hash) == $expected_tree ]] ||
  fail "the copy's files changed during the benchmark, so the runs did not all see the same tree; see $logs"

# A warm and an edit run in the same round should report the same findings; when they do not, the edit
# changed more than bytes.
if measuring warm && measuring edit; then
  awk -F'\t' '$1 == "warm" { warm[$2] = $6 } $1 == "edit" { edit[$2] = $6 }
    END { for (round in edit) if (edit[round] != warm[round])
      printf "note: round %s edit reported %s findings and warm %s, so the edit was not neutral\n", round, edit[round], warm[round] }' $runs_table
fi

# What the primes and warm runs should never do. After the first round the cache holds the last round, so a
# prime that took nothing from it is #547dhjz, a cache the run before left unusable. A warm run is a whole
# replay or the benchmark measured something else under its name. The first round's first prime may start
# cold, after a new engine, and is not warned on.
if [[ -f $logs/primes.tsv ]]; then
  awk -F'\t' '$1 > 1 && $3 == "none" {
      printf "warning: round %s prime %s took nothing from the cache (#547dhjz); see prime-%s-%s.log\n", $1, $2, $1, $2 }' $logs/primes.tsv
fi
measuring warm && awk -F'\t' '$1 == "warm" && $7 != "replay" {
    printf "warning: round %s warm run read cache %s, not a whole replay\n", $2, $7 }' $runs_table

print
print "verdict seconds, by mode (quiet runs are the number; loaded runs are shown, not reported):"
outcome=0
for mode in $modes; do
  table=$(table_of $mode)
  quiet=($(quiet_runs $mode $table))
  loaded=($(awk -F'\t' -v mode=$mode '$1 == mode && $12 == "loaded" { print $3 }' $table | sort -n))
  if (( ${#quiet} > 0 )); then
    printf '  %-5s quiet best %s, median %s, worst %s (%d of %d runs quiet)\n' $mode \
      $quiet[1] $quiet[$(( (${#quiet} + 1) / 2 ))] $quiet[-1] ${#quiet} $rounds_taken
  else
    outcome=3
    printf '  %-5s no quiet number: in all %d runs the machine was under %s%% idle before or after\n' $mode $rounds_taken $idle_floor
  fi
  (( ${#loaded} > 0 )) &&
    printf '        loaded runs, not a quiet number: %s\n' "${(j:, :)loaded}"
done

# The one comparison the first mode exists for (#0q6nmnt), when both it and cold were measured.
if measuring first && measuring cold; then
  first_quiet=($(quiet_runs first $first_table))
  cold_quiet=($(quiet_runs cold $runs_table))
  if (( ${#first_quiet} > 0 && ${#cold_quiet} > 0 )); then
    first_median=$first_quiet[$(( (${#first_quiet} + 1) / 2 ))]
    cold_median=$cold_quiet[$(( (${#cold_quiet} + 1) / 2 ))]
    if awk -v first=$first_median -v cold=$cold_median 'BEGIN { exit !(first > cold * 1.25) }'; then
      printf 'warning: the first run'"'"'s quiet median %s is more than 25%% over cold'"'"'s %s: a first run is doing work cold does not (#0q6nmnt); see first-*.log\n' \
        $first_median $cold_median
      (( outcome == 0 )) && outcome=4
    fi
  fi
fi
print "runs and logs: $logs"

# The record is written from the same runs table the summary above printed, so the two cannot differ, and
# a mode with no quiet run is recorded with no quiet number.
if [[ $record == true ]]; then
  record_flags=(-modes ${(j:,:)${(C)record_modes}} -band $band)
  [[ $stopped_early == true ]] && record_flags+=(-stopped-early)
  measuring edit && record_flags+=(-edit $edit)
  (cd ${0:A:h:h} && go run ./internal/benchresults/tools/record -runs $runs_table -engine $engine \
    -project ${project:t} -project-commit $project_commit -idle-floor $idle_floor -settle $settle \
    -rounds $rounds_taken $record_flags) || fail "could not record the run; its logs are in $logs"
  # The website's benchmarks page is generated from bench/results, and a record committed alone leaves
  # docs/data stale, which TestTheRenderedFilesAreCurrent fails (@system_cohere_web_data).
  print "commit the record with what \`go run ./internal/docsdata/tools/generate\` rewrites in docs/data"
fi
exit $outcome
