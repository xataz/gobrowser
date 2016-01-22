# goBrowser

## What goBrowser

goBrowser is a simple web application for list, download or share yours files.

[![Build Status](https://travis-ci.org/xataz/gobrowser.svg?branch=master)](https://travis-ci.org/xataz/gobrowser)

## Screenshot
![File browser](http://image.noelshack.com/fichiers/2016/01/1452285607-gobrowser-filebrowser.png "File browser")

![Share Listing](http://image.noelshack.com/fichiers/2016/01/1452285613-gobrowser-listshare.png "Share Listing")


## How to install
### Install go
```
$ apt-get install go
```

### Git clone and compile
```
cd /opt
git clone https://github.com/xataz/gobrowser.git
cd gobrowser
go build app.go
```

### Run
Run with default option :
./app

#### Argument
* config="": a string, choose a configfile
* forcessl=false: a bool, force https for share link
* forceurl="": a string, force domain for share link
* hiddenfile=false: a bool, enable hidden files
* listen="127.0.0.1:5000": a string, choose listen port and bind address
* path="/home": a string, choose root path for gobrowser
* webroot="": a string, choose webroot (ex : /files for access with http://mydomain/files)

Example :
```
./app -hiddenfile -listen 0.0.0.0:8080 -path /home/user -webroot /files
```

#### configfile
app.conf.exemple is a example of configfile, run with :
```
./app -config app.conf
```

### Init script
I create an init script for systemd.
Copy it in /lib/systemd/system/gobrowser.service
