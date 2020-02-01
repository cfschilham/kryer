# Installation
There are several ways to install Kryer. The simpelest is probably to download the pre-compiled binaries for your specific system and architecture.

**Pre-compiled Binaries**  
Download the [latest release](https://github.com/cfschilham/kryer/releases/latest) from the releases tab which matches your system. Then copy the binary to `/usr/bin`, after that you will be able to run it using `kryer` in your terminal. Example:  
`$ unzip kryer-v2.0.0-linux-amd64.zip`  
`$ sudo cp kryer /usr/bin/kryer`  
You can now run it: `$ kryer --help`

On Windows you can place the executable in a new folder in Program Files, for example, and then add it to your environment variables. 

To set your environment variables open Control Panel > System and Security > System > Advanced System Settings > Environment Variables

Now select path and click edit, then browse and navigate to the containing folder of the executable. Then simply press OK and you should be able to run it using the `kryer` command in the command prompt.
 
**Building from Source**  
If your pre-compiled binaries are not available for your system or you don't want to use them for other reasons, you can build it yourself from source. To do so you will need a working Go environment. 

Start by cloning the repository into `YourGopath/src/github.com/cfschilham/kryer`. You can then build and install it using `$ sudo make install`, unless you do not have a `/usr/bin` directory.

If that is the case you can build using `$ make build` or `$ go build` in the Kryer directory.  
  
# Usage
To run Kryer, you must always specify at least a dictionary file and a host or hostlist file. Simple, single-threaded attack:  
`$ kryer -h root@192.168.0.0 -d yourdict.txt`  
  
To enable multi-threaded mode, you must specify the maximum amount of concurrent goroutines you would like to have at any given time. You should not go too high on this because your target might not be able to handle, say, 40+ concurrent SSH connection attempts. Example:  
`$ kryer -h root@192.168.0.0 -d yourdict.txt -t 20`  
  
You can also use a list of hosts to connect to instead of a single host. This can also be used to try different usernames on one host. Example:  
`$ kryer -H hostlist.txt -d yourdict.txt -t 20`  
  
With `hostlist.txt` containing, for instance:  
```
root@192.168.0.0
admin@192.168.0.0
user@192.168.0.0
root@192.168.0.6
admin@192.168.0.6
user@192.168.0.6
```
  
Output is also available:  
`$ kryer -h root@192.168.0.0 -d yourdict.txt -t 20 -o outputfile.txt`  
This file will be created if not already present and any found combinations will be written to this file in the following form: `username@adress:password`.
