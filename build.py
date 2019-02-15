import os
import glob
import shutil

# os.environ['GOPATH'] = os.getcwd()

if os.path.isdir('./release'):
    shutil.rmtree("./release")

if not os.path.isdir("./release"):
    os.mkdir("./release")

goos_list = ("linux", "darwin", "freebsd", "windows", )
goarch_list = ("amd64", "386", )
go_list = glob.glob("./*.go")

for goos in goos_list:
    for arch in goarch_list:
        zip_filename = "sectool_%s_%s.zip" % (goos, arch)
        file_list = []
        os.environ['GOOS'] = goos
        os.environ['GOARCH'] = arch
        print(goos, arch)

        for f in go_list:
            if f.endswith(".go"):
                output_file = "%s_%s_%s" % (f[:-3], goos, arch)
                if goos == "windows":
                    output_file = "%s_%s_%s.exe" % (f[:-3], goos, arch)
            cmd = "go build -o {0} {1}".format(output_file, f)
            file_list.append(output_file)
            os.system(cmd)

        zip_cmd = "zip -r %s %s" % (zip_filename, " ".join(file_list))
        rm_cmd = "rm %s" % " ".join(file_list)
        os.system(zip_cmd) 
        os.system(rm_cmd)
        shutil.move(zip_filename, "./release")
