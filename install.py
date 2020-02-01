try:
    import os, subprocess, shutil, stat, json, sys, platform
    from zipfile import ZipFile

    PYTHON2 = 2
    PYTHON3 = 3

    VERSION = 0

    LINUX = 1
    DARWIN = 2
    WINDOWS = 3

    OS = 0


    if sys.version_info[0] < 3:
        import urllib2 as urllib
        VERSION = PYTHON2
    else:
        VERSION = PYTHON3
        import urllib.request as urllib

    if platform.system().lower() == "linux":
        OS = LINUX
    elif platform.system().lower() == "darwin":
        OS = DARWIN
    elif platform.system().lower() == "windows":
        OS = WINDOWS
    else:
        print("Unsupported os " + platform.system().lower())

    print("Getting latest version from https://api.github.com/repos/cfschilham/autossh/releases/latest")
    req = urllib.Request("https://api.github.com/repos/cfschilham/autossh/releases/latest")
    f = urllib.urlopen(req)
    release = json.loads(f.read())
    print("Starting system scan")
    if os.path.isfile('/bin/autossh'):
        if(VERSION == PYTHON2):
            q = raw_input("Autossh already installed do want to reinstall [Y/n]? ")
        if(VERSION == PYTHON3):
            q = str(input("Autossh already installed do want to reinstall [Y/n]? "))

        if(q == "n"):
            exit()

        try:
            os.remove("/bin/autossh")
        except OSError:
            print("Not enough permission to remove /bin/autossh please excecute with sudo!!")
            exit()
        print("Old file removed...")

    print("Getting correct version for your operating system from " + release["assets_url"] + "...")
    req = urllib.Request(release["assets_url"])
    f2 = urllib.urlopen(req)
    versions = json.loads(f2.read())

    for version in versions:
        if(platform.system().lower() in version["name"]):
            print("Downloading " + version["browser_download_url"] + "...")
            if(VERSION == PYTHON2):
                u = urllib.urlopen(version["browser_download_url"])
                datatowrite = u.read()

                with open("/tmp/autossh.zip", 'wb') as f:
                    f.write(datatowrite)

            if(VERSION == PYTHON3):
                urllib.urlretrieve(version["browser_download_url"], "/tmp/autossh.zip")
            break;

    print("Extracting " + version["name"] + "...")
    name = version["name"].replace(".zip", "")
    with ZipFile("/tmp/autossh.zip", 'r') as zipObj:
        if(os.path.isdir('/tmp/autossh') == False):
            os.mkdir('/tmp/autossh')
        zipObj.extractall("/tmp/autossh")

    if(os.path.isdir('/etc/autossh') == False):
        os.mkdir('/etc/autossh')

    if(os.path.isdir('/etc/autossh/config') == False):
        os.mkdir('/etc/autossh/config')
    print("Creating files...")

    shutil.move("/tmp/autossh/" + name + "/autossh", "/bin/autossh")
    shutil.move("/tmp/autossh/" + name + "/cfg/config.yml", "/etc/autossh/config/config.yml")
    shutil.move("/tmp/autossh/" + name + "/cfg/dict.txt", "/etc/autossh/config/dict.txt")
    shutil.move("/tmp/autossh/" + name + "/cfg/hostlist.txt", "/etc/autossh/config/hostlist.txt")
    print("Making executable...")
    st = os.stat('/bin/autossh')
    os.chmod('/bin/autossh', st.st_mode | stat.S_IEXEC)
    print("Cleaning up...")
    shutil.rmtree("/tmp/autossh")
    os.remove("/tmp/autossh.zip")
    print("Done...")
except KeyboardInterrupt:
    print("\nKeyBoardInterrupt detected!!")
    print("Cleaning up...")
    if(os.path.isdir('/tmp/autossh')):
        shutil.rmtree("/tmp/autossh")
    if(os.path.isfile('/tmp/autossh.zip')):
        os.remove("/tmp/autossh.zip")
    print("Exiting...")