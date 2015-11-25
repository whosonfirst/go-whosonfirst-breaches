package breaches

import (
	rtree "github.com/whosonfirst/go-whosonfirst-rtree"
	// polyclip "github.com/akavel/polyclip-go"
	geojson "github.com/whosonfirst/go-whosonfirst-geojson"
)

type Index struct {
	rtree.WOFIndex
}

func (idx *Index) Breaches(feature *geojson.WOFFeature) {

	// do an initial rtree comparison
	// foreach result load feature
	// do polyclip test between feature and result feature

}
