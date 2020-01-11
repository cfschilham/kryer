# Installation  
The simplest way to install is to download the pre-compiled binaries from the releases tab. At present, pre-compiled versions are available for OSX, Linux and Windows using AMD64 architecture.  
  
Alternatively, you can build it yourself using the source files.  
  
**Pre-Compiled Binaries**  
Download the latest release from the [releases tab](https://github.com/cfschilham/autossh/releases) which corresponds to your system.

  
**Build from Source**  
Sometimes, the pre-compiled binaries won't work with CGO, meaning you are unable to resolve .local hostnames. This can typically be fixed by building from source.

Provided you have Go installed and configured properly, you can run `$ go build` inside the repository, which will compile everything into a single file. You can then delete the internal directory and the autossh.go file.
  
# Configuration  
  All configuration can be done inside of cfg/config.yml. 
  
  
|Option|Default|Description|
|--|--|--|
|verbose|false|More verbose output|
|usr_is_host|false|Hostnames/IP's are redundant when this is true, instead they will be drived from the username as: username + .local|
|multi_threaded|true|Enables the use of multiple threads per host. Will spawn goroutines for every password but is capped by max_threads.|
|max_threads|10|Maximum amount of goroutines per host when multi_threaded is true. For example: if max_threads is 10 and the dictionary length is 25, 10 routines will be spawned, after completion that another 10 will be spawned and finally 5 more will be spawned to complete the dictionary.|
|mode|"manual"|Can be set to either "manual" or "hostlist". In manual mode you enter hosts manually one-by-one. In hostlist mode they are read from a hostlist file (separated by newlines).|
|port|"22"|The SSH port to connect to, almost unexpectionally is 22.|
|dict_path|"cfg/dict.txt"|The path of the dictionary file.|
|hostlist_path|"cfg/hostlist.txt"|The path of the hostlist file. Ignored is mode is "manual". With usr_is_host this file should only contain usernames. Otherwise it should contain entries in the form: username@ip each followed by a newline.| 
|output_path|"output.txt"|The path of the output file. This is where all user/password combinations are exported to. If no filename is specified, nothing will be outputted.|