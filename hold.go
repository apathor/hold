package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type HoldDir struct {
	path string
}

func NewHoldDir(path string) (*HoldDir, error) {
	dir := HoldDir{path}
	os.Mkdir(path, 0x700)
	return &dir, nil
}

func (h *HoldDir) validCacheName(name string) bool {
	matched, err := regexp.Match("[[:alnum:]]+", []byte(name))
	if err != nil {
		return false
	}
	return matched
}

func (h *HoldDir) Caches() ([]string, error) {
	// consider all files in the directory
	found, err := filepath.Glob(h.path + "/*")
	if err != nil {
		return nil, err
	}
	// count the files per cache name
	names := make(map[string]int)
	for i := 0; i < len(found); i++ {
		base := filepath.Base(found[i])
		toks := strings.Split(base, ".")
		names[toks[0]]++
	}
	// get a sorted list of cache names
	var keys []string
	for key := range names {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(x, y int) bool { return keys[x] > keys[y] })
	return keys, nil
}

func (h *HoldDir) Files(name string, expiry time.Time) ([]string, []string, error) {
	// find files with this cache prefix
	cache := filepath.Join(h.path, name)
	found, err := filepath.Glob(cache + ".*")
	if err != nil {
		return nil, nil, err
	}
	if len(found) == 0 {
		return nil, nil, errors.New("No cached files.")
	}
	sort.Slice(found, func(x, y int) bool { return found[x] > found[y] })
	// divide files by expiration
	var hot []string
	var cold []string
	for i := 0; i < len(found); i++ {
		// determine filename encoded time
		base := filepath.Base(found[i])
		toks := strings.Split(base, ".")
		ts, err := strconv.Atoi(toks[1])
		if err != nil {
			return nil, nil, err
		}
		htime := time.Unix(int64(ts), 0)
		if expiry.Before(htime) {
			hot = append(hot, found[i])
		} else {
			cold = append(cold, found[i])
		}
	}
	return hot, cold, nil
}

func (h *HoldDir) Stash(name string, input []byte) ([]byte, string, error) {
	now := strconv.Itoa(int(time.Now().Unix()))
	stem := filepath.Join(h.path, name)
	path := strings.Join([]string{stem, now}, ".")
	// TODO lock
	cache, err := os.Create(path)
	if err != nil {
		return nil, "", err
	}
	defer cache.Close()
	_, err = cache.Write(input)
	return input, path, err
}

func (h *HoldDir) Retrieve(name string, expiry time.Time) ([]byte, string, error) {
	hot, _, err := h.Files(name, expiry)
	if err != nil {
		return nil, "", err
	}
	if len(hot) == 0 {
		return nil, "", errors.New("no cached files")
	}
	path := hot[0]
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	return content, path, nil
}

func (h *HoldDir) Load(name string, expiry time.Time, fill func() ([]byte, error)) ([]byte, string, error) {
	// require valid cache name
	if !h.validCacheName(name) {
		return nil, "", errors.New("invalid cache name")
	}
	hot, _, err := h.Files(name, expiry)
	if err != nil {
		return nil, "", err
	}

	var out []byte
	var path string
	// load cached file if possible otherwise populate new file
	if len(hot) > 0 {
		var err error
		out, path, err = h.Retrieve(name, expiry)
		if err != nil {
			return nil, "", err
		}
	} else {
		result, err := fill()
		if err != nil {
			return nil, "", err
		}
		out, path, err = h.Stash(name, result)
		if err != nil {
			return nil, "", err
		}
	}
	return out, path, nil
}

type Cat struct {
	files []string
}

func (c *Cat) Output() ([]byte, error) {
	var out []byte
	for i := 0; i < len(c.files); i++ {
		content, err := ioutil.ReadFile(c.files[i])
		if err != nil {
			return nil, err
		}
		out = append(out, content...)
	}
	return out, nil
}

type HoldArgs struct {
	mode       int
	output     int
	directory  string
	name       string
	expiration time.Time
	command    string
	cmdArgs    []string
	files      []string
}

func GetHoldArgs() (*HoldArgs, error) {
	// accept options
	mode := 0
	modeCommand := flag.Bool("e", false, "arguments are evaluated as a command and the output is cached")
	modeFiles := flag.Bool("f", false, "arguments are considered files and their content is cached")
	modeRetrieve := flag.Bool("g", false, "ignore arguments and only retreive cached content")
	output := 0
	outContent := flag.Bool("p", false, "print the content of the cached file instead of its name")
	outQuiet := flag.Bool("q", false, "be quiet - do not print cache file name or contents")
	dir := flag.String("d", os.Getenv("HOLD_DIR"), "path of hold directory")
	name := flag.String("n", "", "name of cache")
	keep := flag.Duration("t", 0, "expire older than seconds")
	nocache := flag.Bool("x", false, "do not load cached version")
	flag.Parse()
	// program mode has a priority order
	if *modeCommand {
		mode = 0
	}
	if *modeFiles {
		mode = 1
	}
	if *modeRetrieve {
		mode = 2
	}
	// output mode has a priority order
	if *outContent {
		output = 1
	}
	if *outQuiet {
		output = 2
	}
	// a sensible cache directory is used if not specified
	if *dir == "" {
		*dir = filepath.Join(os.Getenv("HOME"), ".cache", "hold")
	}
	// positional arguments might be required
	if flag.NArg() == 0 {
		switch mode {
		case 0:
			return nil, errors.New("expected command")
		case 1:
			return nil, errors.New("expected one or more files")
		}
	}
	files := flag.Args()
	command := flag.Arg(0)
	commandArgs := flag.Args()[1:]
	// require the cache name
	if *name == "" {
		*name = command
	}
	// determine cache expiration time
	var expiration time.Time
	if *nocache {
		expiration = time.Now()
	} else {
		expiration = time.Now().Add(-*keep)
	}
	// success
	new := HoldArgs{
		mode:       mode,
		output:     output,
		directory:  *dir,
		name:       *name,
		expiration: expiration,
		command:    command,
		cmdArgs:    commandArgs,
		files:      files,
	}
	return &new, nil
}

func main() {
	// accept options
	args, err := GetHoldArgs()
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	// setup cache directory
	hold, err := NewHoldDir(args.directory)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	// do it!
	var out []byte
	var path string
	switch args.mode {
	case 0: // cache command output
		cmd := exec.Command(args.command, args.cmdArgs...)
		out, path, err = hold.Load(args.name, args.expiration, cmd.Output)
	case 1: // cache file contents
		cat := Cat{files: args.files}
		out, path, err = hold.Load(args.name, args.expiration, cat.Output)
	case 2: // retreive only
		out, path, err = hold.Retrieve(args.name, args.expiration)
	}
	// report any errors
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	// write desired output
	switch args.output {
	case 0:
		fmt.Printf("%s\n", path)
	case 1:
		fmt.Printf("%s\n", out)
	case 2:
	}
}
