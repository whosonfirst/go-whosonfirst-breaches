package breaches

import (
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
		idx.Logger.Debug("compare %d polys from feature", len(feature_polys))

		for _, feature_poly := range feature_polys {

			clipping, _ := idx.WOFPolygonToPolyclip(&feature_poly.OuterRing)

			for _, r := range inflated {

				// Note the WOFID-iness of this example - please to be waiting for
				// this to get pushed to get resolved (20151125/thisisaaronland)
				// https://github.com/whosonfirst/go-whosonfirst-geojson/issues/2

				if r.Id == feature.WOFId() {
					continue
				}

				result_polys, err := idx.LoadPolygons(r)

				idx.Logger.Debug("Compare %d polys from candidate %d (%s)", len(result_polys), r.Id, r.Name)

				if err != nil {
					idx.Logger.Warning("Unable to load polygons for ID %d, because %v", r.Id, err)
					continue
				}

				for _, p := range result_polys {

					subject, _ := idx.WOFPolygonToPolyclip(&p.OuterRing)
					i := subject.Construct(polyclip.INTERSECTION, *clipping)

					idx.Logger.Debug("%v", i)

					if len(i) > 0 {
						breaches = append(breaches, r)
						break
					}
				}
			}

		}
	}

	return breaches, nil
}

// sudo add me to go-whosonfirst-geojson ?

func (idx *Index) WOFPolygonToPolyclip(p *geo.Polygon) (*polyclip.Polygon, error) {

	poly := new(polyclip.Polygon)
	contour := new(polyclip.Contour)

	for _, _pt := range p.Points() {

		pt := polyclip.Point{
			X: _pt.Lng(),
			Y: _pt.Lat(),
		}

		contour.Add(pt)
	}

	poly.Add(*contour)

	return poly, nil
}
