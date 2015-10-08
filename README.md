# libaursir

Building
========

Prerequisites:

* This libary is currently only supported on Linux, since the go compiler only can not create shared objects on Windows and OSX.
* You will need to install zeromq from your package manager or compile it from source: https://zeromq.com.
* Make sure have at least Go 1.5
* Install aursir4go: ```go get github.com/joernweissenborn/aursir4go```

Compiling:

* clone the repository
* Export: ```go build --buildmode=c-shared -o libaursirexport.so export.go```
* Import: ```go build --buildmode=c-shared -o libaursirimport.so import.go```

Then copy the files to your sytems library folder, e.g. /usr/lib.

Basic usage
===========

Please see https://github.com/joernweissenborn/aursir4py.
