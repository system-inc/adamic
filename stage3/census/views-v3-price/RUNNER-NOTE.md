# Recovered runner naming incident

The first batch stopped before measuring ts.performance.ts because Path.with_suffix
removed the distinguishing dotted suffix in its artifact name. The existing-record
guard prevented overwrite. The three completed records and archives remain intact.
Artifact stems now replace internal dots as well as slashes. The failed attempt
contains no census output and is excluded. runner-naming-incident.tar.gz retains
the precise error. This is resolved; measurement resumes on the remaining files.
