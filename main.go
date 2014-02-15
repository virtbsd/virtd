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
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "github.com/coopernurse/gorp"
    /* "github.com/virtbsd/VirtualMachine" */
    "github.com/virtbsd/jail"
    "github.com/virtbsd/network"
)

var db *gorp.DbMap

func main() {
    db = initDb()
    myjail := jail.GetJail(db, map[string]interface{} { "uuid": "My UUID" })
    fmt.Printf("%+v\n", myjail)
}

func initDb() *gorp.DbMap {
    db, err := sql.Open("mysql", "virtbsd:v1rtbsd!@unix(/tmp/mysql.sock)/virtbsd_db?loc=Local")
    if err != nil {
        return nil
    }

    dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}

    dbmap.AddTable(jail.Jail{}).SetKeys(false, "UUID")
    dbmap.AddTable(network.Network{}).SetKeys(false, "UUID")
    dbmap.AddTable(network.NetworkDevice{}).SetKeys(false, "UUID")
    dbmap.AddTable(network.NetworkPhysical{})
    dbmap.AddTable(network.DeviceAddress{})
    dbmap.AddTable(network.DeviceOption{})

    if err = dbmap.CreateTablesIfNotExists(); err != nil {
        panic(err)
    }

    return dbmap
}
