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
		idx.Logger.Debug("compare %d polys from feature", len(clipping_polys))

		inflated := idx.InflateSpatialResults(results)

		// sudo do this concurrently 

		for _, subject := range inflated {

			if subject.Id == feature.WOFId() {
				continue
			}

			subject_polys, err := idx.LoadPolygons(subject)

			if err != nil {
				idx.Logger.Warning("Unable to load polygons for ID %d, because %v", subject.Id, err)
				continue
			}

			intersects, err := idx.Intersects(clipping_polys, subject_polys)

			if err != nil {
				idx.Logger.Error("Failed to determine intersection, because %v", err)
				continue
			}

			if !intersects {
				continue
			}

			breaches = append(breaches, subject)
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
     

	intersects := false

	for c, clipping_poly := range clipping_polys {

		idx.Logger.Debug("clipping poly %d has %d interior rings", c, len(clipping_poly.InteriorRings))

		clipping_outer, _ := idx.WOFPolygonToPolyclip(&clipping_poly.OuterRing)

		for _, subject := range subject_polys {

			idx.Logger.Debug("subject poly has %d interior rings", len(subject.InteriorRings))

			subject_outer, _ := idx.WOFPolygonToPolyclip(&subject.OuterRing)
			intersection := subject_outer.Construct(polyclip.INTERSECTION, *clipping_outer)

			idx.Logger.Debug("INTERSECTION of clipping (outer poly) and subject (outer poly) %v", len(intersection))

			if len(intersection) > 0 {
				intersects = true
			}
			
			if intersects && len(clipping_poly.InteriorRings) > 0 {

			   intersects = false

			   for i, inner := range clipping_poly.InteriorRings {

			       	    clipping_inner, _ := idx.WOFPolygonToPolyclip(&inner)

					xor := subject_outer.Construct(polyclip.XOR, *clipping_inner)
					idx.Logger.Debug("XOR of clipping (inner poly %d) and subject (outer poly) %d", i, len(xor))

					if len(xor) > 0 {
					   intersects = true
					}
			   }

			}

			if intersects && len(subject.InteriorRings) > 0 {

			   intersects = false

			   for i, inner := range subject.InteriorRings {

			       	    subject_inner, _ := idx.WOFPolygonToPolyclip(&inner)

					xor := clipping_outer.Construct(polyclip.XOR, *subject_inner)
					idx.Logger.Debug("XOR of clipping (outer poly) and subject (inner poly %d) %d", i, len(xor))

					if len(xor) > 0 {
					   intersects = true
					}
			   }

			}
		
			idx.Logger.Debug("does clipping poly %d intersect subject: %t", c, intersects)

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
