# no-shadow-restricted-names on shared checker facts

The port uses the harness checker and shared native reference write analysis. Its descriptor declares no program reads, matching upstream exactly. The typed witnesses include shorthand writes to an undefined binding, an independent unwritten shadow, and reportGlobalThis:false alongside a NaN diagnostic.

The full branch report, blockers, commands, captured inputs, runtime measurements and package logs are in [the wave report](../id-denylist/REPORT.md) and [evidence](../id-denylist/evidence/).

Certification limit: three upstream AllowGlobalThis:true controls lose that flag in shared capture because the Go options field has json:"-". The port supports the configured option, but agreement on their captured {} is not proof of those original three controls. This is named in the wave report; no shared file or capture guard is changed.
