# Hold
Speed up pipelines with caching!

## SYNOPSYS
hold [OPTIONS] [MODE] [ARGUMENTS]

## MODE
 -e          : Arguments are considered a command to be evaludated and the output is cached
 
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

Cache a long running command.

`$ hold -e -n foo -t 1m -- long running command`

Cache a long running pipeline.

`$ hold -q -g -n bar -t 30s || hold -f -x -n bar <(long running command)`

Get a cached file by name.

`$ f=$(hold -g -t 300s -n foo); cat "$f"`

Retreive cached file contents.

`$ hold -g -t 300s -p -n foo`

Overwrite a cache.

`$ hold -f -x -n bar - <<< "test"`

Accept a file argument or stdin in one command.

`$ i=$(hold -f -x -n qux "${1:--}")`


## INSTALL

`go get github.com/apathor/hold`
