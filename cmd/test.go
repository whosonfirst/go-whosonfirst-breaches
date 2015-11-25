package main

import (
       "flag"
       "fmt"
       breaches "github.com/whosonfirst/go-whosonfirst-breaches"
       log "github.com/whosonfirst/go-whosonfirst-log"
       "os"
)

func main() {

     flag.Parse()
     args := flag.Args()

     source := "/usr/local/mapzen/whosonfirst-data/data"

     logger := log.NewWOFLogger("debug")
     logger.AddLogger(os.Stdout, "debug")

     b, _ := breaches.NewIndex(source, 1024, (1024*1024*10), logger)

     fmt.Printf("%v", b)

     for _, path := range args {
          b.IndexGeoJSONFile(path)
     }

}

