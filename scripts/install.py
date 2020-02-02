try:
    import os
    import subprocess
    import shutil
    import stat
    import json
    import sys
    import platform
    import tarfile
    import hashlib

    PYTHON2 = 2
    PYTHON3 = 3

    pyversion = None

    # Detect OS and architecture
    if sys.version_info[0] < 3:
        pyversion = PYTHON2
        import urllib2 as urllib
    else:
        pyversion = PYTHON3
        import urllib.request as urllib

    if platform.machine() != "x86_64":
        if platform.machine() != "":
            print("System architecture unsupported, build from source instead")
            exit()
        elif pyversion == PYTHON3:
            q = input(
                "Unable to detect system architecture, are you using x86_64 (AMD64) [y/N]: ")
            if q.lower() != "y":
                print("System architecture unsupported, build from source instead")
                exit(0)
        else:
            q = raw_input(
                "Unable to detect system architecture, are you using x86_64 (AMD64) [y/N]: ")
            if q.lower() != "y":
                print("System architecture unsupported, build from source instead")
                exit(0)

    if not platform.system() == "Linux" or not platform.system == "Darwin":
        if platform.system == "Windows":
            print(
                "Windows is not supported by the Python installer, use pre-compiled binaries instead.")
        else:
            print("OS unsupported, build from source instead")

    # Check for existance and write permissions
    if not os.path.isdir("/usr/bin"):
        print("Unable to locate /usr/bin")
        exit(1)
    if not os.path.isdir("/tmp"):
        print("Unable to locate /tmp")
        exit(1)

    if not os.access("/usr/bin", os.W_OK):
        print("Insufficient permissions, please elevate")
        exit(1)
    if not os.access("/tmp", os.W_OK):
        print("Insufficient permissions, please elevate")
        exit(1)

    # Get latest release
    print("Getting latest release information from https://api.github.com/repos/cfschilham/kryer/releases/latest ...")
    response = urllib.urlopen(
        "https://api.github.com/repos/cfschilham/kryer/releases/latest")
    release = json.loads(response.read())

    print("Gathering system info ...")
    if os.path.isfile("/usr/bin/kryer"):
        if pyversion == PYTHON2:
            q = raw_input(
                "Kryer is already installed, reinstall/update [Y/n]: ")
        else:
            q = str(
                input("Kryer is already installed, reinstall/update [Y/n]: "))

        if q.lower() == "n":
            exit(0)

        print("Removing /usr/bin/kryer ...")

        try:
            os.remove("/usr/bin/kryer")
        except OSError:
            print("Insufficient permissions, please elevate")
            exit(1)

        print("Removed /usr/bin/kryer ...")

    print("Getting asset list from " + release["assets_url"] + " ...")

    response = urllib.urlopen(release["assets_url"])
    assets = json.loads(response.read())

    for asset in assets:
        if platform.system().lower() in asset["name"] and "tar.gz" in asset["name"]:
            print("Downloading " + asset["browser_download_url"] + " ...")
            if(pyversion == PYTHON2):
                response = urllib.urlopen(asset["browser_download_url"])
                rawfile = response.read()

                with open("/tmp/kryer.tar.gz", "wb") as FILE:
                    FILE.write(rawfile)

            else:
                urllib.urlretrieve(
                    asset["browser_download_url"], "/tmp/kryer.tar.gz")
            break

    print("Extracting " + asset["name"] + " ...")
    filename = asset["name"].replace(".tar.gz", "")
    for asset in assets:
        if asset["name"] == filename + ".sha256":
            print("Getting SHA256 checksum from " +
                  asset["browser_download_url"] + " ...")
            if pyversion == PYTHON2:
                response = urllib.urlopen(asset["browser_download_url"])
                rawfile = response.read()

                with open("/tmp/kryer.sha256", "wb") as FILE:
                    FILE.write(rawfile)

            else:
                urllib.urlretrieve(
                    asset["browser_download_url"], "/tmp/kryer.sha256")
            break

    with open("/tmp/kryer.sha256", "r") as checksum:
        checksum = checksum.read().split(" ")[0]
    print("Verifying checksum " + checksum + " ...")

    checksumFILE = hashlib.sha256()
    with open("/tmp/kryer.tar.gz", "rb") as FILE:
        for chunk in iter(lambda: FILE.read(4096), b""):
            checksumFILE.update(chunk)

    if checksum.strip() != checksumFILE.hexdigest().strip():
        print("Integrity check failed, invalid checksum")
        print("Calculated checksum: " + checksumFILE.hexdigest().strip())
        print("Provided checksum: " + checksum.strip())

        if os.path.isfile("/tmp/kryer"):
            os.remove("/tmp/kryer")
        if os.path.isdir("/tmp/kryer.sha256"):
            shutil.rmtree("/tmp/kryer.sha256")
        if os.path.isfile("/tmp/kryer.tar.gz"):
            os.remove("/tmp/kryer.tar.gz")
        exit(1)

    print("Integrity check successful")

    os.mkdir("/tmp/kryer")
    tar = tarfile.TarFile.open("/tmp/kryer.tar.gz", "r:gz")
    tar.extractall("/tmp/kryer")
    tar.close()

    print("Creating files in /usr/bin ...")

    shutil.move("/tmp/kryer/" + filename + "/kryer", "/usr/bin/kryer")
    print("Making executable ...")

    shutil.rmtree("/tmp/kryer")
    os.remove("/tmp/kryer.tar.gz")
    os.remove("/tmp/kryer.sha256")
    print("Installation successful, use kryer command to start")

except KeyboardInterrupt:
    print("\nKeyboard interrupt detected")
    print("Cleaning up ...")

    if os.path.isdir("/tmp/kryer"):
        shutil.rmtree("/tmp/kryer")
    if os.path.isfile("/tmp/kryer.tar.gz"):
        os.remove("/tmp/kryer.tar.gz")
    if os.path.isfile("/tmp/kryer.sha256"):
        os.remove("/tmp/kryer.sha256")
