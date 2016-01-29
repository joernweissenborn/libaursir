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
	"github.com/joernweissenborn/aurarath/config"
	"github.com/joernweissenborn/aurarath/messages"
	"github.com/joernweissenborn/aurarath/servicedescriptor"
	"github.com/joernweissenborn/aursir4go"
	"github.com/joernweissenborn/eventual2go"
)

var nextImport = 0

var request = map[int]map[string]*eventual2go.Future{}
var listen_results = map[int]*eventual2go.Collector{}
var n2n_streams = map[int]map[string]*eventual2go.Collector{}

func getNextImport() (n int) {
	n = nextImport
	nextImport++
	return
}

var imports = map[int]*aursir4go.Import{}

func newImport(desc *servicedescriptor.ServiceDescriptor, cfg *config.Config) (n int) {
	n = getNextImport()
	i := aursir4go.NewImport(desc, cfg)
	imports[n] = i
	request[n] = map[string]*eventual2go.Future{}
	listen_results[n] = eventual2go.NewCollector()
	listen_results[n].AddStream(i.Results())
	i.Run()
	return
}

//export NewImportYAML
func NewImportYAML(descYAML *C.char, address *C.char) (e C.int) {
	desc := servicedescriptor.FromYaml(C.GoString(descYAML))
	cfg := config.Default(C.GoString(address))
	//	cfg.Logger() = os.Stdout
	return C.int(newImport(desc, cfg))
}

//export Listen
func Listen(i C.int, function *C.char) {
	imp := imports[int(i)]
	if imp != nil {
		imp.Listen(C.GoString(function))
	}
}

//export StopListen
func StopListen(i C.int, function *C.char) {
	imp := imports[int(i)]
	if imp != nil {
		imp.StopListen(C.GoString(function))
	}
}

//export Call
func Call(i C.int, function *C.char, parameter *C.char) *C.char {
	imp := imports[int(i)]
	if imp != nil {
		req := imp.NewRequestBin(C.GoString(function), []byte(C.GoString(parameter)), messages.ONE2ONE)

		isRes := func(uuid string) eventual2go.Filter {
			return func(d eventual2go.Data) bool {
				return d.(*messages.Result).Request.UUID == uuid
			}
		}

		request[int(i)][req.UUID] = imp.Results().FirstWhere(isRes(req.UUID))
		imp.Deliver(req)
		return C.CString(req.UUID)
	}
	return nil
}

//export CallAll
func CallAll(i C.int, function *C.char, parameter *C.char) *C.char {
	imp := imports[int(i)]
	if imp != nil {
		req := imp.NewRequestBin(C.GoString(function), []byte(C.GoString(parameter)), messages.ONE2MANY)

		isRes := func(uuid string) eventual2go.Filter {
			return func(d eventual2go.Data) bool {
				return d.(*messages.Result).Request.UUID == uuid
			}
		}

		n2n_streams[int(i)][req.UUID] = eventual2go.NewCollector()
		n2n_streams[int(i)][req.UUID].AddStream(imp.Results().Where(isRes(req.UUID)))
		imp.Deliver(req)
		return C.CString(req.UUID)
	}
	return nil
}

//export Trigger
func Trigger(i C.int, function *C.char, parameter *C.char) {
	imp := imports[int(i)]
	if imp != nil {
		imp.Deliver(imp.NewRequestBin(C.GoString(function), []byte(C.GoString(parameter)), messages.MANY2ONE))
	}
}

//export TriggerAll
func TriggerAll(i C.int, function *C.char, parameter *C.char) {
	imp := imports[int(i)]
	if imp != nil {
		imp.Deliver(imp.NewRequestBin(C.GoString(function), []byte(C.GoString(parameter)), messages.MANY2MANY))
	}
}

//export GetResult
func GetResult(i C.int, uuid *C.char) *C.char {
	if request[int(i)] == nil {
		return nil
	} else if len(request[int(i)]) == 0 {
		return nil
	} else {
		f := request[int(i)][C.GoString(uuid)]
		if f == nil {
			return nil
		}
		if !f.Completed() {
			return nil
		}
		return C.CString(string(f.GetResult().(*messages.Result).Parameter()))
	}
}

//export GetNextListenResult
func GetNextListenResult(i C.int) *C.char {
	if listen_results[int(i)].Empty() {
		return nil
	} else {
		r := listen_results[int(i)].Preview().(*messages.Result).Request.UUID
		return C.CString(r)
	}
}

//export GetNextListenResultFunction
func GetNextListenResultFunction(i C.int) *C.char {
	if listen_results[int(i)].Empty() {
		return nil
	} else {
		r := listen_results[int(i)].Preview().(*messages.Result).Request.Function
		return C.CString(r)
	}
}

//export GetNextListenResultInParameter
func GetNextListenResultInParameter(i C.int) *C.char {
	if listen_results[int(i)].Empty() {
		return nil
	} else {
		r := string(listen_results[int(i)].Preview().(*messages.Result).Request.Parameter())
		return C.CString(r)
	}
}

//export GetNextListenResultParameter
func GetNextListenResultParameter(i C.int) *C.char {
	if listen_results[int(i)].Empty() {
		return nil
	} else {
		r := string(listen_results[int(i)].Get().(*messages.Result).Parameter())
		return C.CString(r)
	}
}

//export GetNextCallAllResultParameter
func GetNextCallAllResultParameter(i C.int, uuid *C.char) *C.char {
	if n2n_streams[int(i)] == nil {
		return nil
	} else if n2n_streams[int(i)][C.GoString(uuid)] == nil {
		return nil
	} else if n2n_streams[int(i)][C.GoString(uuid)].Empty() {
		return nil
	} else {
		r := string(n2n_streams[int(i)][C.GoString(uuid)].Get().(*messages.Result).Parameter())
		return C.CString(r)
	}
}

func main() {
//	i := NewImportYAML(new(C.char),C.CString("127.0.0.1"))
//	fmt.Println(C.GoString(Call(i,C.CString("SayHello"),C.CString(""))))
}
