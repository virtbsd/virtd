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
)

var db *gorp.DbMap

func main() {
    db = initDb()
    myjail := jail.GetJail(db, map[string]interface{} { "name": "My Name" })
    fmt.Printf("%+v\n", myjail)
}

func initDb() *gorp.DbMap {
    db, err := sql.Open("mysql", "virtbsd:v1rtbsd!@unix(/tmp/mysql.sock)/virtbsd_db?loc=Local")
    if err != nil {
        return nil
    }

    dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}

    dbmap.AddTableWithName(jail.Jail{}, "jail").SetKeys(false, "UUID")

    if err = dbmap.CreateTablesIfNotExists(); err != nil {
        panic(err)
    }
    return dbmap
}
