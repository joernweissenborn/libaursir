//	Copyright (c) 2015 Joern Weissenborn
//
//	This file is part of libaursir.
//
//	Foobar is free software: you can redistribute it and/or modify
//	it under the terms of the GNU General Public License as published by
//	the Free Software Foundation, either version 3 of the License, or
//	(at your option) any later version.
//
//	libaursir is distributed in the hope that it will be useful,
//	but WITHOUT ANY WARRANTY; without even the implied warranty of
//	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//	GNU General Public License for more details.
//
//	You should have received a copy of the GNU General Public License
//	along with libaursir.  If not, see <http://www.gnu.org/licenses/>.

package main

import "C"

import (
	"fmt"
	"github.com/joernweissenborn/aurarath/config"
	"github.com/joernweissenborn/aurarath/messages"
	"github.com/joernweissenborn/aurarath/servicedescriptor"
	"github.com/joernweissenborn/aursir4go"
	"github.com/joernweissenborn/eventual2go"
)

var nextExport = 0

func getNextExport() (n int) {
	n = nextExport
	nextExport++
	return
}

var exports = map[int]*aursir4go.Export{}

func newExport(desc *servicedescriptor.ServiceDescriptor, cfg *config.Config) (n int) {
	n = getNextExport()
	e := aursir4go.NewExport(desc, cfg)
	exports[n] = e
	request[n] = map[string]*messages.Request{}
	waiting_request[n] = []string{}
	e.Requests().Listen(getRequest(n))
	e.Run()
	return
}

var request = map[int]map[string]*messages.Request{}
var waiting_request = map[int][]string{}

func getRequest(n int) eventual2go.Subscriber {
	return func(d eventual2go.Data) {
		r := d.(*messages.Request)
		waiting_request[n] = append(waiting_request[n], r.UUID)
		request[n][r.UUID] = r
	}
}

//export NewExportYAML
func NewExportYAML(descYAML *C.char, address *C.char) (e C.int) {
	fmt.Println(C.GoString(address))
	desc := servicedescriptor.FromYaml(C.GoString(descYAML))
	cfg := config.Default(C.GoString(address))
	//	cfg.Logger() = os.Stdout
	return C.int(newExport(desc, cfg))
}

//export GetNextRequestId
func GetNextRequestId(e C.int) (uuid *C.char) {
	if len(waiting_request[int(e)]) == 0 {
		return
	} else {
		uuid = C.CString(waiting_request[int(e)][0])
		waiting_request[int(e)] = waiting_request[int(e)][1:]
		return
	}
}

//export RetrieveRequestFunction
func RetrieveRequestFunction(e C.int, uuid *C.char) (function *C.char) {
	r := request[int(e)][C.GoString(uuid)]
	if r != nil {
		function = C.CString(r.Function)
	}
	return
}

//export RetrieveRequestParams
func RetrieveRequestParams(e C.int, uuid *C.char) (parameter *C.char) {
	r := request[int(e)][C.GoString(uuid)]
	if r != nil {
		parameter = C.CString(string(r.Parameter()))
	}
	return
}

//export Reply
func Reply(e C.int, uuid *C.char, parameter *C.char) C.int {
	r := request[int(e)][C.GoString(uuid)]
	exp := exports[int(e)]
	if r != nil && exp != nil {
		exp.ReplyEncoded(r, []byte(C.GoString(parameter)))
		delete(request[int(e)], C.GoString(uuid))
	}
	return C.int(0)
}

//export Emit
func Emit(e C.int, function *C.char, inparameter *C.char, outparameter *C.char) C.int {
	exp := exports[int(e)]
	if exp != nil {
		exp.EmitEncoded(
			C.GoString(function),
			[]byte(C.GoString(inparameter)),
			[]byte(C.GoString(outparameter)),
		)
	}
	return C.int(0)
}

func main() {
	//NewExportYAML(new(C.char),new(C.char))
}
