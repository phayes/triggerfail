TriggerFail
===========

Fail a command with an exit status of 1 if a trigger string appears in it's output

```sh
USAGE
  triggerfail "<space-seperated-strings>" [--abort] [-v] <command>

OPTIONS
  -abort=false: Abort a running command if a match is found. If abort is not passed the command is allowed to run to completion
  -v=false: Verbose. Print the reason why we failed the command.

EXAMPLES
  triggerfail "hoot" echo "I don't give a hoot"
  triggerfail --abort -v "error ERROR warning" drush hw-atomload circae
```
