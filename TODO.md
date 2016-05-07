# TODO

- Add checks to ensure all the parts are using the expected/minimum version. Like FileCopy Client+Server, GoPsExec Client+Server
- Can add verification in unit tests to check when we CopyTo and CopyBack the file date stamps agree. This could even be a live copy verification too
- Swap out `github.com/francoishill/afero` for `github.com/spf13/afero` once our changes are merged in and
  + bug [79](https://github.com/spf13/afero/issues/79) was resolved
  + pull request [80](https://github.com/spf13/afero/pull/80) was merged
- Seems like WMIC on windows gives empty `CommandLine` if not running as same user that launched process? But the Processid and some other fields are visible though


# Notes
On windows the command `WMIC PROCESS WHERE Processid=?? GET ExecutablePath /VALUE | MORE +2` will print the full Exe path to the process id