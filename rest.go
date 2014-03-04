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
    "io/ioutil"
    "github.com/virtbsd/jail"
    "github.com/virtbsd/VirtualMachine"
    "github.com/virtbsd/network"
    "github.com/gorilla/mux"
)

type ActionStatus struct {
    Result string
    ErrorMessage string
}

func unmarshal_jail(w http.ResponseWriter, req *http.Request) *jail.JailJSON {
    jailrest := &jail.JailJSON{}
    body, err := ioutil.ReadAll(req.Body)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        status := ActionStatus{Result: "Error", ErrorMessage: err.Error()}
        bytes, _ := json.MarshalIndent(status, "", "    ")
        w.Write(bytes)
        return nil
    }

    req.Body.Close()
    if err = json.Unmarshal(body, jailrest); err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        status := ActionStatus{Result: "Error", ErrorMessage: err.Error()}
        bytes, _ := json.MarshalIndent(status, "", "    ")
        w.Write(bytes)
        return nil
    }

    return jailrest
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

func AddVmHandler(w http.ResponseWriter, req *http.Request) {
    jailrest := &jail.JailJSON{}
    body, err := ioutil.ReadAll(req.Body)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        status := ActionStatus{Result: "Error", ErrorMessage: err.Error()}
        bytes, _ := json.MarshalIndent(status, "", "    ")
        w.Write(bytes)
        return
    }

    req.Body.Close()
    if err = json.Unmarshal(body, jailrest); err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        status := ActionStatus{Result: "Error", ErrorMessage: err.Error()}
        bytes, _ := json.MarshalIndent(status, "", "    ")
        w.Write(bytes)
        return
    }

    obj := &jail.Jail{}
    obj.Name = jailrest.Name
    obj.HostName = jailrest.HostName
    obj.ZFSDataset = jailrest.ZFSDataset
    obj.NetworkDevices = jailrest.NetworkDevices
    obj.Routes = jailrest.Routes
    obj.Options = jailrest.Options

    for _, device := range obj.NetworkDevices {
        for _, address := range device.Addresses {
            address.DeviceAddressID = 0
        }

        for _, option := range device.Options {
            option.DeviceOptionID = 0
        }
    }

    for _, option := range obj.Options {
        option.OptionID = 0
    }

    if err = obj.Persist(db); err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        status := ActionStatus{Result: "Error", ErrorMessage: err.Error()}
        bytes, _ := json.MarshalIndent(status, "", "    ")
        w.Write(bytes)
        return
    }
}

func DeleteVmHandler(w http.ResponseWriter, req *http.Request) {
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
    if err := myjail.Delete(db); err != nil {
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

func UpdateVmHandler(w http.ResponseWriter, req *http.Request) {
    var myjail *jail.Jail
    vars := mux.Vars(req)

    if _, ok := vars["uuid"]; ok {
        myjail = jail.GetJail(db, map[string]interface{} {"uuid": vars["uuid"]})
    } else {
        return
    }

    jailrest := unmarshal_jail(w, req)
    if jailrest == nil {
        return
    }

    /* Diff each of the properties of the jail */
    if len(jailrest.ZFSDataset) > 0 && myjail.ZFSDataset != jailrest.ZFSDataset {
        myjail.ZFSDataset = jailrest.ZFSDataset
    }

    if len(jailrest.Name) > 0 && myjail.Name != jailrest.Name {
        myjail.Name = jailrest.Name
    }

    if len(myjail.NetworkDevices) > 0 && len(jailrest.NetworkDevices) == 0 {
        for _, device := range myjail.NetworkDevices {
            device.Delete(db)
        }

        myjail.NetworkDevices = make([]*network.NetworkDevice, 0)
    } else {
        /* Check for new/updated network devices */
        for _, restdevice := range jailrest.NetworkDevices {
            if mydevice := network.FindDevice(myjail.NetworkDevices, restdevice); mydevice == nil {
                restdevice.UUID = ""
                for _, option := range restdevice.Options {
                    option.DeviceOptionID = 0
                    option.DeviceUUID = ""
                }

                for _, address := range restdevice.Addresses {
                    address.DeviceAddressID = 0
                    address.DeviceUUID = ""
                }

                myjail.NetworkDevices = append(myjail.NetworkDevices, restdevice)
            }
        }

        /* Check for deleted network devices */
        for i := 0; i < len(myjail.NetworkDevices); i++ {
            mydevice := myjail.NetworkDevices[i]

            if restdevice := network.FindDevice(jailrest.NetworkDevices, mydevice); restdevice == nil {
                mydevice.Delete(db)
                copy(myjail.NetworkDevices[i:], myjail.NetworkDevices[i+1:])
                myjail.NetworkDevices[len(myjail.NetworkDevices)-1] = nil
                myjail.NetworkDevices = myjail.NetworkDevices[:len(myjail.NetworkDevices)-1]
                i--
            }
        }

        for _, restdevice := range jailrest.NetworkDevices {
            jaildevice := network.FindDevice(myjail.NetworkDevices, restdevice)

            /* Check addresses */
            for _, address := range restdevice.Addresses {
                if a := network.FindAddress(jaildevice.Addresses, address); a == nil {
                    address.DeviceAddressID = 0
                    jaildevice.Addresses = append(jaildevice.Addresses, address)
                }
            }

            for i := 0; i < len(jaildevice.Addresses); i++ {
                address := jaildevice.Addresses[i]

                if a := network.FindAddress(restdevice.Addresses, address); a == nil {
                    db.Delete(address)
                    copy(jaildevice.Addresses[i:], jaildevice.Addresses[i+1:])
                    jaildevice.Addresses[len(jaildevice.Addresses)-1] = nil
                    jaildevice.Addresses = jaildevice.Addresses[:len(jaildevice.Addresses)-1]
                    i--
                }
            }

            /* Check options */
            for _, option := range restdevice.Options {
                if o := network.FindOption(jaildevice.Options, option); o == nil {
                    option.DeviceOptionID = 0
                    jaildevice.Options = append(jaildevice.Options, option)
                }
            }

            for i := 0; i < len(jaildevice.Options); i++ {
                option := jaildevice.Options[i]

                if o := network.FindOption(restdevice.Options, option); o == nil {
                    db.Delete(option)
                    copy(jaildevice.Options[i:], jaildevice.Options[i+1:])
                    jaildevice.Options[len(jaildevice.Options)-1] = nil
                    jaildevice.Options = jaildevice.Options[:len(jaildevice.Options)-1]
                    i--
                }
            }
        }
    }
}

func StartRESTService() {
    r := mux.NewRouter()
    r.HandleFunc("/vmapi/1/vm/uuid/{uuid}/status", StatusHandler).Methods("GET")
    r.HandleFunc("/vmapi/1/vm/uuid/{uuid}/start", StartHandler).Methods("GET")
    r.HandleFunc("/vmapi/1/vm/uuid/{uuid}/stop", StopHandler).Methods("GET")
    r.HandleFunc("/vmapi/1/vm/uuid/{uuid}/delete", DeleteVmHandler).Methods("GET")
    r.HandleFunc("/vmapi/1/vm/uuid/{uuid}/update", UpdateVmHandler).Methods("POST")
    r.HandleFunc("/vmapi/1/vm/list", ListHandler).Methods("GET")
    r.HandleFunc("/vmapi/1/vm/add", AddVmHandler).Methods("POST")
    http.Handle("/", r)
    http.ListenAndServe(":9000", nil)
}
