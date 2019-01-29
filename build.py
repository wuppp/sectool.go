import os
import shutil

os.environ['GOPATH'] = os.getcwd()

if not os.path.isdir("./release"):
    os.mkdir("./release")

goos_list = ("linux", "darwin", "freebsd", "windows", )
goarch_list = ("amd64", "386", )
go_list = ("portscan.go", "httpbanner.go", "ping.go", "ssh.go", )

for goos in goos_list:
    for arch in goarch_list:
        filename = "sectool_%s_%s.zip" % (goos, arch)
        print(goos, arch)
        for f in go_list:
            os.environ['GOOS'] = goos
            os.environ['GOARCH'] = arch
            cmd = "go build %s" % f
            os.system(cmd)
        file_list = " ".join([i[:-3] for i in go_list])
        zip_cmd = "zip -r %s %s" % (filename, file_list)
        rm_cmd = "rm %s" % file_list
        if goos == "windows":
            file_list = " ".join([i[:-3] + ".exe" for i in go_list])
            zip_cmd = "zip -r %s %s" % (filename, file_list)
            rm_cmd = "rm %s" % file_list
        os.system(zip_cmd) 
        os.system(rm_cmd)
        shutil.move(filename, "./release")