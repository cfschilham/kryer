# Installation
There are several ways to install Kryer. Pre-compiled binaries are available for Windows, Linux and Darwin with AMD64 architecture. On Darwin and Linux, a Python installer is available.
  
**Python Installer**  
For Darwin and Linux, a Python installer is available. Simply run:  
`$ sh -c "$(curl https://raw.githubusercontent.com/cfschilham/kryer/master/scripts/install.sh -s)"`  
The correct binary will automatically be installed into `/usr/bin`.  
  
**Pre-compiled Binaries**  
Download the [latest release](https://github.com/cfschilham/kryer/releases/latest) from the releases tab which matches your system. Then copy the binary to `/usr/bin`, after that you will be able to run it using `kryer` in your terminal. Example:  
`$ tar -xvzf kryer-v2.0.0-linux-amd64.tar.gz`  
`$ sudo cp kryer-v2.0.0-linux-amd64/kryer /usr/bin/kryer`  
You can now run it:  
`$ kryer --help`  
  
On Windows you can place the executable in a new directory in Program Files, for example, and then add it to your environment variables. 
  
To set your environment variables open Control Panel > System and Security > System > Advanced System Settings > Environment Variables
  
Now select path and click edit, then click browse and select the containing directory of the executable. Press OK and you should be able to run it using the `kryer` command in the command prompt.
  
**Building from Source**  
If pre-compiled binaries are not available for your system or you don't want to use them for other reasons, you can build Kryer yourself from source. To do so you will need a working Go environment. 
  
Start by cloning the repository into `YourGopath/src/github.com/cfschilham/kryer`. You can then build and install it using `$ sudo make install`, unless you do not have a `/usr/bin` directory.
  
If that is the case you can build using `$ make build` or `$ go build` in the Kryer directory.  
  
# Usage
To run Kryer, you must always specify at least a dictionary file and a host or hostlist file. Simple, single-threaded attack:  
`$ kryer -h root@192.168.0.0 -d yourdict.txt`  
  
To enable multi-threaded mode, you must specify the maximum amount of concurrent outgoing connection attempts. You should not set this too high as a remote host might not be able to handle a large amount of concurrent incoming SSH connections. However, to decrease the amount of time it takes to go through a dictionary, it is recommended to use more than 1 (the default). Any number up to 10 should not cause trouble. Numbers up to 40 might be stable but it is recommeded you expirement with this first to avoid skipped dictionary entries due to overload. Example:  
`$ kryer -h root@192.168.0.0 -d yourdict.txt -t 20`  
  
You can also use a list of hosts to connect to instead of a single host. This can also be used to try different usernames on the same host. Example:  
`$ kryer -H hostlist.txt -d yourdict.txt`  
  
Where `hostlist.txt` contains, for instance:  
```
root@192.168.0.0
admin@192.168.0.0
user@192.168.0.0
root@192.168.0.6
admin@192.168.0.6
user@192.168.0.6
```
  
Output to a file is also possible:  
`$ kryer -h root@192.168.0.0 -d yourdict.txt -o outputfile.txt`  
A file will be created if one does not already exist and any found combinations will be written to this file in the following form: `username@address:password`.
