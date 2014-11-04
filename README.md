TriggerFail
===========

Fail a command with an exit status of 1 if a trigger string appears in it's output. This is incredibly useful for use with testing framworks (such as Travis CI), which will let you fail a test if certain keywords like "Error" or "Warning" appears in it's output.

```sh
USAGE
  triggerfail "<space-seperated-strings>" [--abort] [-v] <command>

OPTIONS
  -abort=false: Abort a running command if a match is found. If abort is not passed the command is allowed to run to completion
  -v=false: Verbose. Print the reason why we failed the command.

EXAMPLE
  triggerfail --abort -v "Error Warning" mysqldump my_database > mysqlbackup.sql #Abort a running mysqldump if we encounter a warning or error.
```
