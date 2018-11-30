#!/bin/bash
version=$1
if [ $# -eq 0 ] 
then
        echo "Please input version, like \"./release.sh 0.60\""
        exit
fi
rm -f release/sectool_*$1.tgz
echo "Build ReleaseFile for version $version"

export GOPATH=`pwd`

echo "build linux_amd64"
export GOOS=linux GOARCH=amd64 
go build portscan.go
go build httpbanner.go
tar zcvf sectool_linux_x64_$1.tgz portscan httpbanner
rm -f portscan httpbanner portscan.exe httpbanner.exe

echo "build linux_386"
export GOOS=linux GOARCH=386 
go build portscan.go
go build httpbanner.go
tar zcvf sectool_linux_x86_$1.tgz portscan httpbanner
rm -f portscan httpbanner portscan.exe httpbanner.exe

echo "build mac_x64"
export GOOS=darwin GOARCH=amd64 
go build portscan.go
go build httpbanner.go
tar zcvf sectool_mac_x86_$1.tgz portscan httpbanner
rm -f portscan httpbanner portscan.exe httpbanner.exe

echo "build win32"
export GOOS=windows GOARCH=386 
go build portscan.go
go build httpbanner.go
tar zcvf sectool_win32_$1.tgz portscan.exe httpbanner.exe
rm -f portscan httpbanner portscan.exe httpbanner.exe

echo "build win64"
export GOOS=windows GOARCH=amd64 
go build portscan.go
go build httpbanner.go
tar zcvf sectool_win64_$1.tgz portscan.exe httpbanner.exe
rm -f portscan httpbanner portscan.exe httpbanner.exe

echo "Build Over"

mkdir release
mv *.tgz release
ls -l release/sectool_*$1.tgz