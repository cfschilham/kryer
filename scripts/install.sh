#!/bin/bash

if [ `command -v python` != "" ]; then
    sudo python -c "$(curl https://raw.githubusercontent.com/cfschilham/kryer/master/scripts/install.py -s)"
elif [ `command -v python3` != "" ]; then
    sudo python3 -c "$(curl https://raw.githubusercontent.com/cfschilham/kryer/master/scripts/install.py -s)"
else
    echo "Please install any version of python"
fi
