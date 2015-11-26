package main

import (
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
	var loglevel = flag.String("loglevel", "debug", "...")
	var subject_id = flag.Int("subject", 0, "...")
	var clipping_id = flag.Int("clipping", 0, "...")

	flag.Parse()

	logger := log.NewWOFLogger("[wof-breaches] ")
	logger.AddLogger(os.Stdout, *loglevel)

	idx, _ := breaches.NewIndex(*data, *cache_size, *cache_trigger, logger)

	subject_path := utils.Id2AbsPath(*data, *subject_id)
	clipping_path := utils.Id2AbsPath(*data, *clipping_id)

	idx.IndexGeoJSONFile(subject_path)

	subject, _ := idx.LoadGeoJSON(subject_path)
	clipping, _ := idx.LoadGeoJSON(clipping_path)

	logger.Info("does %s (clipping) breach %s (subject)", clipping.WOFName(), subject.WOFName())

	t1 := time.Now()

	results, _ := idx.Breaches(clipping)

	if len(results) == 0 {
		logger.Info("%s (clipping) DOES NOT breach %s (subject)", clipping.WOFName(), subject.WOFName())
	}

	for _, r := range results {

		subject_path := utils.Id2AbsPath(*data, r.Id)
		subject_feature, _ := idx.LoadGeoJSON(subject_path)

		logger.Info("%s (%d) breaches %s (%d)", clipping.WOFName(), clipping.WOFId(), subject_feature.WOFName(), subject_feature.WOFId())
	}

	t2 := time.Since(t1)
	logger.Info("time to search %v", t2)

}
