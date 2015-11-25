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

func (idx *Index) Breaches(feature *geojson.WOFFeature) {

	// do an initial rtree comparison
	// foreach result load feature
	// do polyclip test between feature and result feature

}
