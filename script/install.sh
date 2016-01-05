#!/bin/bash


## VAR
VERSION_GO=1.5.2
OS=linux
[[ $(uname -m) == "x86_64" ]] && ARCH=amd64 || ARCH=386
PWD_TMP=$(pwd)

## check root
if [ $(who) != "root" ]; then
	echo "You must root for install gobrowser"
	exit 1
fi

## Place tmp
cd /tmp

## check if go
which go > /dev/null 2>&1

if [ $? -eq 1 ]; then
	wget https://storage.googleapis.com/golang/go${VERSION_GO}.${OS}-${ARCH}.tar.gz	
	tar -C /usr/local -xzf go${VERSION_GO}.${OS}-${ARCH}.tar.gz
	export PATH=$PATH:/usr/local/go/bin
	export GOROOT=/usr/local/go
	echo "export PATH=\$PATH:/usr/local/go/bin" >> ~/.profile
	echo "export GOROOT=/usr/local/go" >> ~/.profile
else
	echo "go is already install"

fi

cd $PWD_TMP && cd ..

go build app.go

if [ $? -ne 0 ]; then
	echo "Compilation Error"
	exit 1
fi

cp app /usr/local/bin/gobrowser

mkdir -p /usr/etc/gobrowser
cp -r templates /usr/etc/gobrowser/templates
cp app.conf /usr/etc/gobrowser/app.conf


## Check if systemd
which systemctl > /dev/null 2>&1

if [ $? -eq 0 ]; then
	cp script/init.systemd /lib/systemd/system/gobrowser.service
fi
