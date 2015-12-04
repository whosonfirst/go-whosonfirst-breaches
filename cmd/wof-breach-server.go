package main

import (
	"encoding/json"
	"flag"
	"fmt"
	breaches "github.com/whosonfirst/go-whosonfirst-breaches"
	log "github.com/whosonfirst/go-whosonfirst-log"
	utils "github.com/whosonfirst/go-whosonfirst-utils"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

func main() {

	/*
	   This still lacks metrics collection of strict placetype checking
	   (20151203/thisisaaronland)
	*/

	var port = flag.Int("port", 9988, "The port number to listen for requests on")
	var data = flag.String("data", "", "The data directory where WOF data lives, required")
	var cache_size = flag.Int("cache_size", 1024, "The number of WOF records with large geometries to cache")
	var cache_trigger = flag.Int("cache_trigger", 2000, "The minimum number of coordinates in a WOF record that will trigger caching")
	// var strict = flag.Bool("strict", false, "Enable strict placetype checking")
	var logs = flag.String("logs", "", "Where to write logs to disk")
	// var metrics = flag.String("metrics", "", "Where to write (@rcrowley go-metrics style) metrics to disk")
	// var format = flag.String("metrics-as", "plain", "Format metrics as... ? Valid options are \"json\" and \"plain\"")
	var cors = flag.Bool("cors", false, "Enable CORS headers")
	var loglevel = flag.String("loglevel", "info", "")

	flag.Parse()
	args := flag.Args()

	if *data == "" {
		panic("missing data")
	}

	_, err := os.Stat(*data)

	if os.IsNotExist(err) {
		panic("data does not exist")
	}

	l_writer := io.MultiWriter(os.Stdout)

	if *logs != "" {

		l_file, l_err := os.OpenFile(*logs, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)

		if l_err != nil {
			panic(l_err)
		}

		l_writer = io.MultiWriter(os.Stdout, l_file)
	}

	logger := log.NewWOFLogger("[wof-breach-server] ")
	logger.AddLogger(l_writer, *loglevel)

	idx, _ := breaches.NewIndex(*data, *cache_size, *cache_trigger, logger)

	for _, meta := range args {
		logger.Info("indexing %s", meta)
		idx.IndexMetaFile(meta)
	}

	logger.Info("indexing meta files complete")

	var lookups int64
	lookups = 0

	handler := func(rsp http.ResponseWriter, req *http.Request) {

		query := req.URL.Query()

		str_wofid := query.Get("wofid")

		if str_wofid == "" {
			http.Error(rsp, "Missing wofid parameter", http.StatusBadRequest)
			return
		}

		wofid, err := strconv.Atoi(str_wofid)

		if err != nil {
			http.Error(rsp, "Unable to parse WOF ID", http.StatusBadRequest)
			return
		}

		path := utils.Id2AbsPath(*data, wofid)
		feature, err := idx.LoadGeoJSON(path)

		if err != nil {
			http.Error(rsp, "Invalid WOF ID", http.StatusBadRequest)
			return
		}

		logger.Info("looking up %d (%d)", feature.WOFId(), lookups)

		t1 := time.Now()
		atomic.AddInt64(&lookups, 1)

		results, err := idx.Breaches(feature)

		atomic.AddInt64(&lookups, -1)
		t2 := time.Since(t1)

		logger.Info("time to lookup %d : %v", feature.WOFId(), t2)

		if err != nil {
			http.Error(rsp, err.Error(), http.StatusInternalServerError)
			return
		}

		logger.Info("%d results breach %d", len(results), feature.WOFId())

		js, err := json.Marshal(results)

		if err != nil {
			http.Error(rsp, err.Error(), http.StatusInternalServerError)
			return
		}

		if *cors {
			rsp.Header().Set("Access-Control-Allow-Origin", "*")
		}

		rsp.Header().Set("Content-Type", "application/json")
		rsp.Write(js)
	}

	str_port := fmt.Sprintf(":%d", *port)

	http.HandleFunc("/", handler)
	http.ListenAndServe(str_port, nil)
}
