// A stand-in for internal/buildcache: a product directory lives in the build cache, outside every package.
package buildcache

func Product(name string) string { return "/var/cache/adamic/" + name }
