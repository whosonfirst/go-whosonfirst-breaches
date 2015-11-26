package breaches

import (
	polyclip "github.com/akavel/polyclip-go"
	geo "github.com/kellydunn/golang-geo"
	geojson "github.com/whosonfirst/go-whosonfirst-geojson"
	log "github.com/whosonfirst/go-whosonfirst-log"
	rtree "github.com/whosonfirst/go-whosonfirst-rtree"
	"time"
)

type Index struct {
	*rtree.WOFIndex
}

type Intersection struct {
	subject    *geojson.WOFSpatial
	intersects bool
}

func NewIndex(source string, cache_size int, cache_trigger int, logger *log.WOFLogger) (*Index, error) {

	r, err := rtree.NewIndex(source, cache_size, cache_trigger, logger)

	if err != nil {
		return nil, err
	}

	idx := Index{r}

	return &idx, nil
}

// return the list of places (subjects) that are breached by feature (clipping)

func (idx *Index) Breaches(feature *geojson.WOFFeature) ([]*geojson.WOFSpatial, error) {

	clipping, err := feature.EnSpatialize()

	if err != nil {
		return nil, err
	}

	breaches := make([]*geojson.WOFSpatial, 0)

	results := idx.GetIntersectsByRect(clipping.Bounds())

	if len(results) > 0 {

		clipping_polys := feature.GeomToPolygons()
		idx.Logger.Debug("compare %d polys from clipping against %d possible subjects", len(clipping_polys), len(results))

		inflated := idx.InflateSpatialResults(results)

		ch := make(chan Intersection, len(inflated))
		pending := 0

		for _, subject := range inflated {

			if subject.Id == feature.WOFId() {
				continue
			}

			go func(clipping_polys []*geojson.WOFPolygon, subject *geojson.WOFSpatial) {

				subject_polys, err := idx.LoadPolygons(subject)

				if err != nil {
					idx.Logger.Warning("Unable to load polygons for ID %d, because %v", subject.Id, err)
					ch <- Intersection{nil, false}
				}

				intersects, err := idx.Intersects(clipping_polys, subject_polys)

				if err != nil {
					idx.Logger.Error("Failed to determine intersection, because %v", err)
					ch <- Intersection{nil, false}
				}

				if err != nil {
					ch <- Intersection{nil, false}
				}

				ch <- Intersection{subject, intersects}

			}(clipping_polys, subject)

			pending += 1
		}

		for j := 0; j < pending; j++ {

			i := <-ch

			if i.intersects {
				breaches = append(breaches, i.subject)
			}
		}

	}

	return breaches, nil
}

func (idx *Index) Intersects(clipping_polys []*geojson.WOFPolygon, subject_polys []*geojson.WOFPolygon) (bool, error) {

	/*
		         AND (intersection) - create regions where both subject and clip polygons are filled
		    	 OR (union) - create regions where either subject or clip polygons (or both) are filled
			 NOT (difference) - create regions where subject polygons are filled except where clip polygons are filled
		    	 XOR (exclusive or) - create regions where either subject or clip polygons are filled but not where both are filled
	*/

	// https://godoc.org/github.com/akavel/polyclip-go#Op

	intersects := false

	for c, clipping_poly := range clipping_polys {

		idx.Logger.Debug("TEST clipping poly %d (which has %d interior rings) against %d subject polys", c, len(clipping_poly.InteriorRings), len(subject_polys))

		clipping_outer, _ := idx.WOFPolygonToPolyclip(&clipping_poly.OuterRing)

		ch := make(chan bool, len(subject_polys))

		for s, subject_poly := range subject_polys {

			go func(subject_poly *geojson.WOFPolygon, clipping_poly *geojson.WOFPolygon, clipping_outer *polyclip.Polygon, c int, s int) {

				t1 := time.Now()

				_intersects := false

				idx.Logger.Debug("TEST subject poly %d against clipping poly %d", s, c)

				subject_outer, _ := idx.WOFPolygonToPolyclip(&subject_poly.OuterRing)
				intersection := subject_outer.Construct(polyclip.INTERSECTION, *clipping_outer)

				// idx.Logger.Debug("INTERSECTION of clipping (outer poly) %d and subject (outer poly) %d: %v", c, s, len(intersection))

				if len(intersection) > 0 {
					_intersects = true
				}

				if _intersects && len(subject_poly.InteriorRings) > 0 {

					_intersects = false

					for _, inner := range subject_poly.InteriorRings {

						subject_inner, _ := idx.WOFPolygonToPolyclip(&inner)

						xor := clipping_outer.Construct(polyclip.XOR, *subject_inner)
						// idx.Logger.Debug("XOR of clipping (outer poly %d) and subject (inner poly %d:%d) %d", c, s, i, len(xor))

						if len(xor) > 0 {
							_intersects = true
						}
					}

				}

				if _intersects && len(clipping_poly.InteriorRings) > 0 {

					_intersects = false

					for _, inner := range clipping_poly.InteriorRings {

						clipping_inner, _ := idx.WOFPolygonToPolyclip(&inner)

						xor := subject_outer.Construct(polyclip.XOR, *clipping_inner)
						// idx.Logger.Debug("XOR of clipping (inner poly %d:%d) and subject (outer poly %d) %d", c, i, s, len(xor))

						if len(xor) > 0 {
							_intersects = true
						}
					}

				}

				t2 := time.Since(t1)
				idx.Logger.Debug("TIME to calculate intersection for clipping poly %d / subject poly %d : %v (%t)", c, s, t2, _intersects)

				ch <- _intersects

			}(subject_poly, clipping_poly, clipping_outer, c, s)

		}

		possible := len(subject_polys)
		var iters int

		for iters = 0; iters < possible; iters++ {

			_intersects := <-ch

			if _intersects {
				intersects = true
				break
			}
		}

		idx.Logger.Debug("does clipping poly %d intersect subject (%d/%d iterations): %t", c, iters, possible, intersects)

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
