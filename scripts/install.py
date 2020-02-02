try:
    import os, subprocess, shutil, stat, json, sys, platform, tarfile, hashlib

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
    else:
        print("Unsupported os " + platform.system().lower())

    print("Getting latest version from https://api.github.com/repos/cfschilham/kryer/releases/latest")
    req = urllib.Request("https://api.github.com/repos/cfschilham/kryer/releases/latest")
    f = urllib.urlopen(req)
    release = json.loads(f.read())
    print("Starting system scan")
    if os.path.isfile('/usr/bin/kryer'):
        if(VERSION == PYTHON2):
            q = raw_input("kryer already installed do want to reinstall [Y/n]? ")
        if(VERSION == PYTHON3):
            q = str(input("kryer already installed do want to reinstall [Y/n]? "))

        if(q == "n"):
            exit()

        try:
            os.remove("/usr/bin/kryer")
        except OSError:
            print("Not enough permission to remove /bin/kryer please excecute with sudo!!")
            exit()
        print("Removed /usr/bin/kryer...")

    print("Getting correct version for your operating system from " + release["assets_url"] + "...")
    req = urllib.Request(release["assets_url"])
    f2 = urllib.urlopen(req)
    versions = json.loads(f2.read())

    for version in versions:
        if(platform.system().lower() in version["name"] and 'tar.gz' in version["name"]):
            print("Downloading " + version["browser_download_url"] + "...")
            if(VERSION == PYTHON2):
                u = urllib.urlopen(version["browser_download_url"])
                datatowrite = u.read()

                with open("/tmp/kryer.tar.gz", 'wb') as f:
                    f.write(datatowrite)

            if(VERSION == PYTHON3):
                urllib.urlretrieve(version["browser_download_url"], "/tmp/kryer.tar.gz")
            break;

    print("Extracting " + version["name"] + "...")
    name = version["name"].replace(".tar.gz", "")
    for version in versions:
        if(version["name"] == name + ".sha256"):
            print("Getting sha256 from " + version["browser_download_url"] + "...")
            if(VERSION == PYTHON2):
                u = urllib.urlopen(version["browser_download_url"])
                datatowrite = u.read()

                with open("/tmp/kryer.sha256", 'wb') as f:
                    f.write(datatowrite)

            if(VERSION == PYTHON3):
                urllib.urlretrieve(version["browser_download_url"], "/tmp/kryer.sha256")
    
    with open("/tmp/kryer.sha256", "r") as sha256_file:
        checksum = sha256_file.read().split(" ")[0]
    print("Checking sha256 sum " + checksum + "...")

    hash_sha256 = hashlib.sha256()
    with open("/tmp/kryer.tar.gz", "rb") as f:
        for chunk in iter(lambda: f.read(4096), b""):
            hash_sha256.update(chunk)

    if(str(checksum.strip()) != str(hash_sha256.hexdigest().strip())):
        print hash_sha256.hexdigest()
        print("Integrity check failed checksum was not valid...")
        print("Cleaning up...")
        if(os.path.isfile('/tmp/kryer')):
            os.remove("/tmp/kryer")
        if(os.path.isdir('/tmp/kryer.sha256')):
            shutil.rmtree("/tmp/kryer.sha256")
        if(os.path.isfile('/tmp/kryer.tar.gz')):
            os.remove("/tmp/kryer.tar.gz")
        print("Exiting...")
        exit()
    print("Integrity check completed checksum verified...")
    if(os.path.isdir('/tmp/kryer') == False):
        os.mkdir('/tmp/kryer')
    tar = tarfile.TarFile.open("/tmp/kryer.tar.gz", "r:gz")
    tar.extractall("/tmp/kryer")
    tar.close()

    print("Creating files...")

    shutil.move("/tmp/kryer/" + name + "/kryer", "/usr/bin/kryer")
    print("Making executable...")
    st = os.stat('/usr/bin/kryer')
    os.chmod("/usr/bin/kryer", st.st_mode | stat.S_IEXEC)
    print("Cleaning up...")
    shutil.rmtree("/tmp/kryer")
    os.remove("/tmp/kryer.tar.gz")
    os.remove(__file__)
    os.remove("/tmp/kryer.sha256")
    print("Done...")
except KeyboardInterrupt:
    print("\nKeyBoardInterrupt detected!!")
    print("Cleaning up...")
    if(os.path.isdir('/tmp/kryer')):
        shutil.rmtree("/tmp/kryer")
    if(os.path.isfile('/tmp/kryer.tar.gz')):
        os.remove("/tmp/kryer.tar.gz")
    print("Exiting...")

# 528bfa3abd4daf159e10bdc79fd0cb75efd9d4e6665a4a7c4f21c77f540331c5
# 528bfa3abd4daf159e10bdc79fd0cb75efd9d4e6665a4a7c4f21c77f540331c5