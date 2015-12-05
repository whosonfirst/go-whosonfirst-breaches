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
	bounds := clipping.Bounds()
	results := idx.GetIntersectsByRect(bounds)

	// idx.Logger.Info("possible results for %v : %d", bounds, len(results))

	if len(results) > 0 {

		t1 := time.Now()

		clipping_polys := feature.GeomToPolygons()
		idx.Logger.Debug("compare %d polys from clipping against %d possible subjects", len(clipping_polys), len(results))

		// See what's going on here? We are *not* relying on the standard Inflate method
		// but rather bucketing all the WOFSpatials by their WOF ID (this is because the
		// (WOF) rtree package indexes the bounding boxes of individual polygons on the
		// geom rather than the bounding box of the set of polygons) which we we will
		// loop over below (20151130/thisisaaronland)

		inflated := make(map[int][]*geojson.WOFSpatial)

		for _, r := range results {

			wof := r.(*geojson.WOFSpatial)
			wofid := wof.Id

			_, ok := inflated[wofid]

			possible := make([]*geojson.WOFSpatial, 0)

			if ok {
				possible = inflated[wofid]
			}

			possible = append(possible, wof)
			inflated[wofid] = possible
		}

		for wofid, possible := range inflated {

			if wofid == feature.WOFId() {
				idx.Logger.Debug("%d can not breach itself, skipping", wofid)
				continue
			}

			// Despite the notes about goroutines and yak-shaving below this is probably
			// a pretty good place to do things concurrently (21051130/thisisaaronland)

			subject_polys, err := idx.LoadPolygons(possible[0])

			if err != nil {
				idx.Logger.Warning("Unable to load polygons for ID %d, because %v", wofid, err)
				continue
			}

			idx.Logger.Debug("testing %d with %d possible candidates", wofid, len(possible))

			// Note to self: it turns out that goroutine-ing these operations is yak-shaving
			// and often slower (20151130/thisisaaronland)

			for i, subject := range possible {

				idx.Logger.Debug("testing %d (offset %d) with candidate %d", wofid, subject.Offset, i)

				test_polys := make([]*geojson.WOFPolygon, 0)

				if subject.Offset == -1 {
					test_polys = subject_polys
				} else {
					test_polys = append(test_polys, subject_polys[subject.Offset])
				}

				intersects, err := idx.Intersects(clipping_polys, test_polys)

				if err != nil {
					idx.Logger.Error("Failed to determine intersection, because %v", err)
					continue
				}

				if intersects {
					breaches = append(breaches, subject)
					idx.Logger.Debug("determined that %d breaches after %d/%d iterations", wofid, i, len(possible))
					break
				}

			}
		}

		t2 := time.Since(t1)
		idx.Logger.Debug("time to test %d possible results: %v", len(results), t2)
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

			go func(subject_poly *geojson.WOFPolygon, clipping_poly *geojson.WOFPolygon, clipping_outer *polyclip.Polygon, c int, s int, ch chan bool) {

				t1 := time.Now()

				_intersects := false

				idx.Logger.Debug("TEST subject poly %d against clipping poly %d", s, c)

				subject_outer, _ := idx.WOFPolygonToPolyclip(&subject_poly.OuterRing)
				intersection := subject_outer.Construct(polyclip.INTERSECTION, *clipping_outer)

				idx.Logger.Debug("INTERSECTION of clipping (outer poly) %d and subject (outer poly) %d: %v", c, s, len(intersection))

				if len(intersection) > 0 {
					_intersects = true
				}

				if _intersects && len(subject_poly.InteriorRings) > 0 {

					_intersects = false

					for i, inner := range subject_poly.InteriorRings {

						subject_inner, _ := idx.WOFPolygonToPolyclip(&inner)

						xor := clipping_outer.Construct(polyclip.XOR, *subject_inner)
						idx.Logger.Debug("XOR of clipping (outer poly %d) and subject (inner poly %d:%d) %d", c, s, i, len(xor))

						if len(xor) > 0 {
							_intersects = true
						}
					}

				}

				if _intersects && len(clipping_poly.InteriorRings) > 0 {

					_intersects = false

					for i, inner := range clipping_poly.InteriorRings {

						clipping_inner, _ := idx.WOFPolygonToPolyclip(&inner)

						xor := subject_outer.Construct(polyclip.XOR, *clipping_inner)
						idx.Logger.Debug("XOR of clipping (inner poly %d:%d) and subject (outer poly %d) %d", c, i, s, len(xor))

						if len(xor) > 0 {
							_intersects = true
						}
					}

				}

				t2 := time.Since(t1)
				idx.Logger.Debug("TIME to calculate intersection for clipping poly %d / subject poly %d : %v (%t)", c, s, t2, _intersects)

				ch <- _intersects

			}(subject_poly, clipping_poly, clipping_outer, c, s, ch)

		}

		possible := len(subject_polys)
		var iters int

		for iters = 0; iters < possible; iters++ {

			select {
			case _intersects := <-ch:

				if _intersects {
					intersects = true
					break
				}

			// This is a band-aid until we can find a proper solution for this:
			// https://github.com/whosonfirst/go-whosonfirst-breaches/issues/1

			case <-time.After(5 * time.Second):
				idx.Logger.Warning("TIMEOUT determining whether clipping poly %d intersect subject (%d/%d iterations): %t", c, iters, possible, intersects)
				break
			}
		}

		idx.Logger.Debug("DOES clipping poly %d intersect subject (%d/%d iterations): %t", c, iters, possible, intersects)

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
