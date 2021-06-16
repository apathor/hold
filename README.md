# Hold
Speed up pipelines with caching!

## SYNOPSYS
hold [OPTIONS] [MODE] [ARGUMENTS]

## MODE
 -e         : arguments are evaluated as a command and the output is cached
 
 -f         : arguments are considered files and their content is cached
 
 -g         : ignore arguments and only retreive cached content

## OPTIONS
 -n NAME    : Set the cache name. By default the name of the calling function is used.
 
 -p         : Print the contents of the cached file instead of its name.
 
 -q         : Quiet mode. Do not print cache file name or contents.
 
 -t SECONDS : Clear cache files older than given number of seconds.
 
 -x         : Do not load a cached file.

## ENVIRONMENT

HOLD_DIR   : Cache directory. The default is ~/.cache/hold.

## EXAMPLES

Cache a long running command.
`$ hold -e -n foo -t 1m -- long running command`

Cache a long running pipeline.
`$ hold -q -g -n bar -t 30s || hold -f -x -n bar <(long running command)`

Get a cached file by name.
`$ f="\$(hold -g -t 300s -n foo)"; cat "\$f"`

Retreive cache file contents.
`$ hold -g -t 300s -p -n foo`

Overwrite a cache.
`$ hold -f -x -n bar - <<< "test")`

Accept a file argument or stdin in one command.
`$ hold -f -x -n qux "\${1:--}"`
