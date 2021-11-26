# Hold
Speed up pipelines with caching!

## SYNOPSYS
hold [OPTIONS] [MODE] [ARGUMENTS]

## MODE
 -e          : Arguments are considered a command to be evaluated and the output is cached
 
 -f          : Arguments are considered files and their content is cached
 
 -g          : Ignore arguments and only retreive cached content

## OPTIONS
 -n NAME     : Set the cache name
 
 -p          : Print the contents of the cached file instead of its name
 
 -q          : Quiet mode. Do not print cache file name or contents
 
 -t DURATION : Do not accept cache files older than the given duration
 
 -x          : Do not load a cached file

## ENVIRONMENT

HOLD_DIR   : Directory in which to store cache files. The default is ~/.cache/hold.

## EXAMPLES

Cache the output of a command for one minute.

`$ hold -e -p -n foo -t 1m -- date`

Cache the output of a pipeline for thirty seconds.

`$ hold -g -p -n bar -t 30s || hold -f -p -x -n bar < <(sleep 10; date)`

Get a cache file by name.

`$ f=$(hold -g -n baz); cat "$f"`

Retreive cache file contents newer than three-hundred seconds.

`$ hold -g -t 300s -p -n qux`

Overwrite a cache.

`$ hold -f -x -n cor - <<< "test"`

Handle a file argument that reads stdin by default.

`$ cfile() { f=$(hold -f -x -n cfile "${1:--}"); printf "%s\n" "$f"; };`

## INSTALL

`go install github.com/apathor/hold@latest`
