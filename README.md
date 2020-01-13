# Installation  
The simplest way to install is to download the pre-compiled binaries from the releases tab. At present, pre-compiled versions are available for OSX, Linux and Windows using AMD64 architecture.  
  
Alternatively, you can build it yourself using the source files.  
  
**Pre-Compiled Binaries**  
Download the latest release from the [releases tab](https://github.com/cfschilham/autossh/releases) which corresponds to your system.

  
**Build from Source**  
Sometimes, the pre-compiled binaries won't work with CGO, meaning you are unable to resolve .local hostnames. This can typically be fixed by building from source.

Provided you have Go installed and configured properly, you can run `$ make release` inside the repository directory, which will create a new folder with the built files. This folder does not depend on the source code to be present.
  
# Configuration  
  All configuration can be done inside of cfg/config.yml. 
  
  
|Option|Default|Description|
|--|--|--|
|usr_is_host|false|Hostnames/IP's are redundant when this is true, instead they will be derived from the username as: username + .local|
|multi_threaded|true|Enables the use of multiple threads per host. Will continuously assign passwords to a goroutine in the pool of the configured size (until the end of the dictionary is reached, of course).|
|goroutines|10|Goroutine pool size in multi-threaded mode.|
|mode|"manual"|Can be set to either "manual" or "hostlist". In manual mode you enter hosts manually one-by-one. In hostlist mode they are read from a hostlist file (separated by newlines).|
|port|"22"|The SSH port to connect to, almost unexceptionally is 22.|
|dict_path|"cfg/dict.txt"|The path of the dictionary file.|
|hostlist_path|"cfg/hostlist.txt"|The path of the hostlist file. Ignored if mode is "manual". With usr_is_host this file should only contain usernames. Otherwise it should contain entries in the form: username@ip each followed by a newline.|
|output_path|"output.txt"|The path of the output file. Found credentials will be stored here. Set to "" to turn off.|
