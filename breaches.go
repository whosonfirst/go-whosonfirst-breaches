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

		feature_polys := feature.GeomToPolygons()
		idx.Logger.Debug("compare %d polys from feature", len(feature_polys))

		inflated := idx.InflateSpatialResults(results)

		for _, candidate := range inflated {

			if candidate.Id == feature.WOFId() {
				continue
			}

			intersects, err := idx.Intersects(feature_polys, candidate)

			if err != nil {
				idx.Logger.Error("Failed to determine intersection, because %v", err)
				continue
			}

			if !intersects {
				continue
			}

			breaches = append(breaches, candidate)
		}
	}

	return breaches, nil
}

func (idx *Index) Intersects(feature_polys []*geojson.WOFPolygon, candidate *geojson.WOFSpatial) (bool, error) {

	// p.OuterRing
	// p.InteriorRings

	candidate_polys, err := idx.LoadPolygons(candidate)

	if err != nil {
		idx.Logger.Warning("Unable to load polygons for ID %d, because %v", candidate.Id, err)
		return false, err
	}

	intersects := false

	for _, feature_poly := range feature_polys {

		clipping, _ := idx.WOFPolygonToPolyclip(&feature_poly.OuterRing)

		for _, c := range candidate_polys {

			subject, _ := idx.WOFPolygonToPolyclip(&c.OuterRing)
			i := subject.Construct(polyclip.INTERSECTION, *clipping)

			idx.Logger.Debug("INTERSECTION %v", len(i))

			if len(i) > 0 {
				intersects = true
			}

			// If the candidate does intersect check to make sure that the
			// candidate is not also contained by any of the interior rings

			if intersects && len(c.InteriorRings) > 0 {

				intersects = false

				idx.Logger.Debug("TEST INTERIOR RINGS FOR CANDIDATE %d", len(c.InteriorRings))

				for _, inner := range c.InteriorRings {

					inner_subject, _ := idx.WOFPolygonToPolyclip(&inner)

					i := inner_subject.Construct(polyclip.INTERSECTION, *clipping)
					idx.Logger.Debug("INTERSECTION %d", len(i))

					if len(i) > 0 {
						intersects = true
						break
					}
				}
			}

			if intersects {
				break
			}
		}

		if intersects {
			break
		}
	}

	return intersects, nil
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
