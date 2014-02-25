/*
(BSD 2-clause license)

Copyright (c) 2014, Shawn Webb
All rights reserved.

Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

   * Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

package main

import (
    /*
    "log"
    "time"
    */
    "fmt"
    "encoding/json"
    "net/http"
    "github.com/virtbsd/jail"
    "github.com/virtbsd/VirtualMachine"
    "github.com/gorilla/mux"
)

type ActionStatus struct {
    Result string
    ErrorMessage string
}

func StartHandler(w http.ResponseWriter, req *http.Request) {
    var myjail *jail.Jail
    myjail = nil

    vars := mux.Vars(req)

    if _, ok := vars["uuid"]; ok {
        myjail = jail.GetJail(db, map[string]interface{} {"uuid": vars["uuid"]})
    }

    if myjail == nil {
        w.WriteHeader(http.StatusNotFound)
        return
    }

    w.Header().Add("Content-Type", "application/json")

    status := ActionStatus{}
    if err := myjail.PrepareHostNetworking(); err == nil {
        if err = myjail.Start(); err == nil {
            if err = myjail.PrepareGuestNetworking(); err != nil {
                w.WriteHeader(http.StatusInternalServerError)
                status.Result = "Error"
                status.ErrorMessage = err.Error()
            } else {
                if err = myjail.PostStart(); err != nil {
                    w.WriteHeader(http.StatusInternalServerError)
                    status.Result = "Error"
                    status.ErrorMessage = err.Error()
                }
            }
        } else {
            w.WriteHeader(http.StatusInternalServerError)
            status.Result = "Error"
            status.ErrorMessage = err.Error()
        }
    } else {
        w.WriteHeader(http.StatusInternalServerError)
        status.Result = "Error"
        status.ErrorMessage = err.Error()
    }

    if len(status.Result) == 0 {
        status.Result = "Okay"
        w.WriteHeader(http.StatusOK)
    }

    if bytes, err := json.Marshal(&status); err == nil {
        w.Write(bytes)
    } else {
        fmt.Printf("Could not marshal status object: %s\n", err.Error())
    }
}

func StopHandler(w http.ResponseWriter, req *http.Request) {
    var myjail *jail.Jail
    myjail = nil

    vars := mux.Vars(req)

    if _, ok := vars["uuid"]; ok {
        myjail = jail.GetJail(db, map[string]interface{} {"uuid": vars["uuid"]})
    }

    if myjail == nil {
        w.WriteHeader(http.StatusNotFound)
        return
    }

    w.Header().Add("Content-Type", "application/json")

    status := ActionStatus{}
    if err := myjail.Stop(); err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        status.Result = "Error"
        status.ErrorMessage = err.Error()
    } else {
        w.WriteHeader(http.StatusOK)
        status.Result = "Okay"
    }

    if bytes, err := json.Marshal(status); err == nil {
        w.Write(bytes)
    }
}

func StatusHandler(w http.ResponseWriter, req *http.Request) {
    var myjail *jail.Jail
    myjail = nil

    vars := mux.Vars(req)

    if _, ok := vars["uuid"]; ok {
        fmt.Printf("UUID passed in: %s\n", vars["uuid"])
        myjail = jail.GetJail(db, map[string]interface{} {"uuid": vars["uuid"]})
    }

    if myjail == nil {
        w.WriteHeader(http.StatusNotFound)
        return
    }

    w.Header().Add("Content-Type", "application/json")

    if bytes, err := json.MarshalIndent(myjail, "", "    "); err == nil {
        w.Write(bytes)
    } else {
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Printf("Error in marshaling: %s\n", err.Error())
    }
}

type VirtualMachines struct {
    VirtualMachines []VirtualMachine.VirtualMachine
}

func ListHandler(w http.ResponseWriter, req *http.Request) {
    var vms VirtualMachines

    for _, j := range jail.GetAllJails(db) {
        vms.VirtualMachines = append(vms.VirtualMachines, j)
    }

    w.Header().Add("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)

    if bytes, err := json.MarshalIndent(vms, "", "    "); err == nil {
        w.Write(bytes)
        return
    }
}

func StartRESTService() {
    r := mux.NewRouter()
    r.HandleFunc("/vmapi/1/vm/uuid/{uuid}/status", StatusHandler).Methods("GET")
    r.HandleFunc("/vmapi/1/vm/uuid/{uuid}/start", StartHandler).Methods("GET")
    r.HandleFunc("/vmapi/1/vm/uuid/{uuid}/stop", StopHandler).Methods("GET")
    r.HandleFunc("/vmapi/1/vm/list", ListHandler).Methods("GET")
    http.Handle("/", r)
    http.ListenAndServe(":9000", nil)
}
