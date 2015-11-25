package breaches

import (
	"errors"
	polyclip "github.com/akavel/polyclip-go"
	geo "github.com/kellydunn/golang-geo"
	geojson "github.com/whosonfirst/go-whosonfirst-geojson"
	log "github.com/whosonfirst/go-whosonfirst-log"
	rtree "github.com/whosonfirst/go-whosonfirst-rtree"
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

	breaches := make([]*geojson.WOFSpatial, 0)

	results := idx.GetIntersectsByRect(sp.Bounds())

	if len(results) > 0 {

		inflated := idx.InflateSpatialResults(results)

		// p.OuterRing
		// p.InteriorRings

		feature_polys := feature.GeomToPolygons()

		clipping, _ := idx.WOFPolygonToPolyclip(feature_polys.OuterRing)

		for _, r := range inflated {

			result_polys, err := idx.LoadPolygons(r)

			if err != nil {
				idx.Logger.Warning("...")
				continue
			}

			for _, p := range result_polys {

				subject, _ := idx.WOFPolygonToPolyclip(p.OuterRing)

				r := subject.Construct(polyclip.INTERSECTION, clipping)

				// account for interior rings
			}
		}
	}

	return breaches, nil
}

func (idx *Index) WOFPolygonToPolyclip(*geo.Polygon) (polyclip.Polygon, error) {
	return nil, errors.New("please write me")
}
