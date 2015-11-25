package main

import (
       "fmt"
       breaches "github.com/whosonfirst/go-whosonfirst-breaches"
       log "github.com/whosonfirst/go-whosonfirst-log"
       "os"
)

func main() {

     source := "/usr/local/mapzen/whosonfirst-data/data"

     logger := log.NewWOFLogger("debug")
     logger.AddLogger(os.Stdout, "debug")

     b := new(breaches.Index)
     idx := b.NewIndex(source, 1024, (1024*1024*10), logger)

     fmt.Printf("%v", idx)
}

