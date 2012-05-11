watch_ninja
===========

watch_ninja combines the awesome ninja build system with the power of inotify to rebuild
source files as you edit them.

Installing
----------

```go get github.com/scottfranklin/watch_ninja```

(Binary distributions of Go1 do not include the inotify library.
If you are using one, the necessary files will be fetched from github.com/dersebi/go_exp.)

Usage
-----

```
cd your/ninja/project
watch_ninja
```

watch_ninja will rebuild the first target that depends on any source file you modify, and
it will automatically keep its source list up to date as you modify build.ninja.

Known Issues
------------

Automatic reload of source file list does not play nicely with subninjas, though if your build.ninja
is always the last thing to be modified, then things should work ok.