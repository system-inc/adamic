# What the benchmarks in this directory share: a copy of a project pinned to its commits, the one
# cohere engine every run uses, the machine's load, and a timer that waits for a run's last process.
# Sourced, not run. The sourcing script sets `name` (for its messages), `project`, `work`, `cohere`,
# `ceiling` and `settle` first.
#
# The copy. The project's committed tree is extracted (its submodules too, each at the commit the
# parent records) into the work directory, and the project's node_modules are linked in rather than
# copied. So the input is named by commits, not by whatever was on disk, nothing another session is
# editing can move a number, and the project's own caches are never read or written. A copy is reused
# by later invocations at the same commits, and refused if its files no longer match what was extracted.
#
# The copy keeps the checkout's repository shape, because cohere reads it: a `.git` file in each submodule
# makes it a nested repository the format walk skips and discovery treats as linked, and the root's `.git`
# directory carries the project's own `info/exclude`, which the format walk applies. Without them the copy
# formatted libraries/structure as part of ahra, 5,269 files where a developer's run formats 3,198
# (#679s763). git itself never runs on the copy, so the markers hold nothing else.
#
# The engine. The launcher rebuilds whenever a cohere commit lands, which during a working session is
# every few minutes, so a benchmark run through it can measure two binaries and report one number. The
# engine the launcher resolves at the start is copied beside the logs and run directly.
zmodload zsh/datetime

fail() { print -u2 "$name: $*"; exit 1; }
load() { sysctl -n vm.loadavg | awk '{print $2}'; }
under_ceiling() { awk -v load=$1 -v ceiling=$ceiling 'BEGIN { exit !(load <= ceiling) }'; }
# idle_over prints the share of all cores idle, in percent, over the next <seconds>. It is iostat's second
# report, which covers that interval (the first covers the time since boot), read from the end of the line,
# where the three load averages follow the cpu columns whatever disks are listed before them.
idle_over() { iostat -c 2 -w $1 | tail -1 | awk '{ print $(NF - 3) }'; }
# wait_for_quiet waits up to `settle` seconds for the load to reach the ceiling, and returns either way.
wait_for_quiet() {
  (( settle > 0 )) || return 0
  local deadline=$(( EPOCHREALTIME + settle ))
  until under_ceiling $(load) || (( EPOCHREALTIME > deadline )); do sleep 5; done
}

# extract writes the tree of <repository> at <commit> into <destination>, then each submodule at the
# commit that tree records. A submodule commit the local clone does not have stops the benchmark rather
# than leaving a hole in the program. Each submodule is marked a repository of its own, as a checkout's is.
extract() {
  local repository=$1 commit=$2 destination=$3
  mkdir -p $destination
  git -C $repository archive $commit | tar -x -C $destination || fail "extracting $repository at $commit"
  [[ $destination != $copy ]] &&
    print "gitdir: none, a benchmark copy of $repository at $commit" > $destination/.git
  git -C $repository ls-tree -r $commit | awk '$2 == "commit" { print $3 "\t" $4 }' |
    while IFS=$'\t' read -r submodule_commit submodule_path; do
      git -C $repository/$submodule_path cat-file -e "$submodule_commit^{commit}" 2> /dev/null ||
        fail "$repository/$submodule_path does not have $submodule_commit, which $commit records"
      print -u2 "  submodule $submodule_path at ${submodule_commit[1,12]}"
      local within=${destination#$copy}
      print "${within#/}${within:+/}$submodule_path at $submodule_commit" >> $copy/bench-submodules
      extract $repository/$submodule_path $submodule_commit $destination/$submodule_path
    done
}

# tree_hash is one hash over every file in the copy, the caches and linked node_modules aside, so a
# copy that moved by a byte is caught.
tree_hash() {
  (cd $copy && find . -type f -not -path './.cache/*' -not -path './bench-ready' -print0 | sort -z |
    xargs -0 shasum -a 256 | shasum -a 256 | awk '{print $1}')
}

# prepare_copy sets project_commit, copy and expected_tree, extracting the copy if this commit has none.
prepare_copy() {
  project=${project:A}
  [[ -d $project/.git || -f $project/.git ]] || fail "$project is not a git checkout"
  project_commit=$(git -C $project rev-parse HEAD) || fail "cannot read $project's HEAD"
  # The layout is in the name, so a copy extracted before the repository markers is never reused.
  copy=$work/${project:t}-${project_commit[1,12]}-repositories
  if [[ -f $copy/bench-ready ]]; then
    print -u2 "reusing the copy at $copy"
    [[ $(tree_hash) == $(< $copy/bench-ready) ]] ||
      fail "the copy at $copy no longer matches what was extracted; remove it and run again"
  else
    [[ -e $copy ]] && fail "$copy exists but was never finished; remove it and run again"
    print -u2 "extracting ${project:t} at ${project_commit[1,12]} into $copy"
    extract $project $project_commit $copy
    # The root is a repository, as the project's checkout is, with the same local excludes.
    mkdir -p $copy/.git/info
    local exclude=$(git -C $project rev-parse --path-format=absolute --git-path info/exclude)
    [[ -f $exclude ]] && cp $exclude $copy/.git/info/exclude
    # Every package in the copy uses the project's installed dependencies, linked rather than copied.
    (cd $copy && find . -name package.json -not -path '*/node_modules/*') | while read -r manifest; do
      local package=${manifest:h}
      [[ -d $project/$package/node_modules ]] && ln -s $project/$package/node_modules $copy/$package/node_modules
    done
    tree_hash > $copy/bench-ready
  fi
  expected_tree=$(< $copy/bench-ready)
}

# pin_engine sets cohere_version, engine_source and engine, a copy of the engine in <logs>. Asking the
# launcher for its version is what makes it build the engine for the current commit, so the newest engine
# in its binary cache is that one, and the two versions agreeing proves it.
#
# Another session's launcher can build a newer commit between the two reads, so a mismatch is asked again,
# three times, before it is refused.
pin_engine() {
  local logs=$1 attempt
  whence -p $cohere > /dev/null || [[ -x $cohere ]] || fail "no cohere at $cohere"
  local launcher=$(whence -p $cohere || print ${cohere:A})
  engine=$logs/cohere-engine
  for attempt in 1 2 3; do
    cohere_version=$(cd $copy && $cohere --version 2>&1) || fail "cohere --version failed: $cohere_version"
    if [[ ${launcher:A:t} == cohere-dispatch ]]; then
      local newest=(${launcher:A:h}/cohere-darwin-*(N.om[1]))
      (( ${#newest} == 1 )) || fail "the launcher's binary cache at ${launcher:A:h} holds no engine"
      engine_source=$newest[1]
    else
      engine_source=${launcher:A}
    fi
    cp $engine_source $engine
    [[ $(cd $copy && $engine --version 2>&1) == $cohere_version ]] && return 0
    print -u2 "the newest engine is not the build the launcher reports, asking again"
    sleep 5
  done
  fail "the engine at $engine_source is not the build the launcher reports; run again once nothing is building"
}

# print_pinned_header prints the commits, the machine and the engine every report starts with.
print_pinned_header() {
  print "# project  ${project:t} at $project_commit"
  [[ -f $copy/bench-submodules ]] && sed 's/^/#   submodule /' $copy/bench-submodules
  print "# machine  $(sysctl -n hw.model), $(sysctl -n machdep.cpu.brand_string), $(sysctl -n hw.ncpu) cores," \
    "$(( $(sysctl -n hw.memsize) / 1073741824 )) GB, macOS $(sw_vers -productVersion)"
  print "# cohere   $engine_source"
  print -- "$cohere_version" | sed 's/^/#   /'
  print "# copy     $copy"
}

# run_timed runs a command in the copy and sets verdict_seconds, settled_seconds and exit_code. It does
# what the cohere launcher does, hands the command a descriptor to write its exit code to and takes the
# verdict when that byte arrives, and then waits for the command's whole process group, so a cache
# write still running cannot overlap the next run. A command that never writes the byte has its
# verdict when it exits, so for any tool but cohere the two times are the same run to its end.
run_timed() {
  local log=$1 timing=$1.timing; shift
  (cd $copy && perl -MPOSIX -MTime::HiRes=time -e '
    my $timing = shift;
    pipe(my $reader, my $writer) or die "pipe: $!";
    my $started = time;
    my $engine = fork() // die "fork: $!";
    if ($engine == 0) {
      setpgrp(0, 0);
      POSIX::dup2(fileno($writer), 3) or die "dup2: $!";
      $ENV{COHERE_VERDICT_FD} = 3;
      exec(@ARGV) or die "exec: $!";
    }
    close $writer;
    my $exit_code;
    $exit_code = ord($verdict) if sysread($reader, my $verdict, 1) == 1;
    my $answered = time;
    waitpid($engine, 0);
    $exit_code //= $? >> 8;
    select(undef, undef, undef, 0.02) while kill(0, -$engine);
    my $settled = time;
    open(my $out, ">", $timing) or die "$timing: $!";
    printf $out "%.3f %.3f %d\n", $answered - $started, $settled - $started, $exit_code;
  ' $timing "$@") < /dev/null > $log 2>&1 || fail "could not run $1, see $log"
  read -r verdict_seconds settled_seconds exit_code < $timing
}
