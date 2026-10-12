"""Where stage3 keeps its caches when no variable names one: where location.BuildCache keeps Go's products, so a runner
that names its build cache gets nothing written in its home directory, which may be read-only or shared (#ka4qkfk)."""
import os
import sys
from pathlib import Path


def cache_root():
    """The directory stage3's caches live under: ADAMIC_BUILD_CACHE_DIR when it is set, refused under ADAMIC_SANDBOX=1
    (where every location is declared), else the user cache directory as Go's os.UserCacheDir names it."""
    configured = os.environ.get('ADAMIC_BUILD_CACHE_DIR')
    if configured:
        return Path(configured).resolve()
    if os.environ.get('ADAMIC_SANDBOX') == '1':
        raise RuntimeError('ADAMIC_SANDBOX=1 keeps caches only where ADAMIC_BUILD_CACHE_DIR or STAGE3_CACHE says, '
                           'and neither is set')
    if sys.platform == 'darwin':
        return Path.home() / 'Library' / 'Caches'
    xdg = os.environ.get('XDG_CACHE_HOME')
    if xdg and Path(xdg).is_absolute():
        return Path(xdg)
    return Path.home() / '.cache'


def stage3_cache():
    """STAGE3_CACHE when it is set, else adamic-stage3 under cache_root()."""
    configured = os.environ.get('STAGE3_CACHE')
    return Path(configured) if configured else cache_root() / 'adamic-stage3'
