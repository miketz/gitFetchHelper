# git fetch helper for sub modules

This program is for a specific case of fetching upstream remotes of git sub modules in my Emacs config.
Asynchronously fetching all of them at once. This is much faster than fetching 1 by 1.

In my emacs config I moved away from git submodules (slow on windows).
Now I just use normal clones into an ignored (by git) directory "notElpaYolo/".
All functions in this project are still relevant and work the same.
But several extra helper functions have been added to support this style of package management as
the usual way of getting latest on .emacs.d/
```bash
git pull
git submodule update --init
```
no longer gets any of the repos in the ignored folder notElpaYolo/.


# how to build 

ideally use make
```bash
make
```

or if you are on windows with no make command just use the Go tooling directly
```bash
go build
```

or use -ldflags to omit symbol table, debug info, and dwarf symbol table. (smaller binary).
```bash
go build -o gitFetchHelper -ldflags="-s -w"
```