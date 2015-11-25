package breaches

import (
	rtree "github.com/whosonfirst/go-whosonfirst-rtree"
	// polyclip "github.com/akavel/polyclip-go"
	geojson "github.com/whosonfirst/go-whosonfirst-geojson"
	log "github.com/whosonfirst/go-whosonfirst-log"
)

type Index struct {
	*rtree.WOFIndex
}

func NewIndex(source string, cache_size int, cache_trigger int, logger *log.WOFLogger) (*Index, error) {

	r, err := rtree.NewIndex(source, cache_size, cache_trigger, logger)

	if err != nil {
		return nil, err
	}

	idx := Index{r}

	return &idx, nil
}

func (idx *Index) Breaches(feature *geojson.WOFFeature) ([]*geojson.WOFSpatial, error) {

	sp, err := feature.EnSpatialize()

	if err != nil {
		return nil, err
	}

	bounds := sp.Bounds()

	results := idx.GetIntersectsByRect(bounds)
	inflated := idx.InflateSpatialResults(results)

	breaches := make([]*geojson.WOFSpatial, 0)

	// construct Polyclip thing-y from feature here

	for _, r := range inflated {

		polys, err := idx.LoadPolygons(r)

		if err != nil {
			idx.Logger.Warning("...")
			continue
		}

		for _, p := range polys {

			idx.Logger.Debug("%v", p)

			// construct Polyclip thing-y from feature here

			// compare Polyclip thing-ies here

			// account for interior rings
		}
	}

	return breaches, nil
}
