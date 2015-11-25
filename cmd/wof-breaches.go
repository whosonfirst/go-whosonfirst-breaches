package main

import (
	"bufio"
	"flag"
	breaches "github.com/whosonfirst/go-whosonfirst-breaches"
	log "github.com/whosonfirst/go-whosonfirst-log"
	utils "github.com/whosonfirst/go-whosonfirst-utils"
	"os"
	"time"
)

func main() {

	var data = flag.String("data", "", "The data directory where WOF data lives, required")
	var cache_size = flag.Int("cache_size", 1024, "The number of WOF records with large geometries to cache")
	var cache_trigger = flag.Int("cache_trigger", 2000, "The minimum number of coordinates in a WOF record that will trigger caching")
	var loglevel = flag.String("loglevel", "info", "...")

	flag.Parse()
	args := flag.Args()

	logger := log.NewWOFLogger("[wof-breaches] ")
	logger.AddLogger(os.Stdout, *loglevel)

	idx, _ := breaches.NewIndex(*data, *cache_size, *cache_trigger, logger)

	for _, meta := range args {
		logger.Info("indexing %s", meta)
		idx.IndexMetaFile(meta)
	}

	logger.Info("indexing meta files complete")

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {

		path := scanner.Text()

		if path == "" {
			continue
		}

		feature, err := idx.LoadGeoJSON(path)

		if err != nil {

			logger.Error("Failed to read %s, because %v", path, err)
			continue
		}

		t1 := time.Now()

		results, err := idx.Breaches(feature)

		if err != nil {
			logger.Error("Failed to determined breaches for %s (%d), because %v", feature.WOFName(), feature.Id, err)
			continue
		}

		for _, r := range results {

			other_path := utils.Id2AbsPath(*data, r.Id)
			other_feature, _ := idx.LoadGeoJSON(other_path)

			logger.Info("%s (%d) breaches %s (%d)", feature.WOFName(), feature.Id, other_feature.WOFName(), other_feature.Id)
		}

		t2 := time.Since(t1)
		logger.Info("time to search %v", t2)
	}

}
