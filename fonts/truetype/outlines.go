package truetype

// this file converts from font format for glyph outlines to
// segments that rasterizer will consume
//
// adapted from snft/truetype.go

type SegmentOp uint8

const (
	SegmentOpMoveTo SegmentOp = iota
	SegmentOpLineTo
	SegmentOpQuadTo
	SegmentOpCubeTo
)

type SegmentPoint struct {
	X, Y float32 // in fonts units
}

type Segment struct {
	Op SegmentOp
	// Args is up to three (x, y) coordinates, depending on the
	// operation.
	// The Y axis increases down.
	Args [3]SegmentPoint
}

func midPoint(p, q SegmentPoint) SegmentPoint {
	return SegmentPoint{
		X: (p.X + q.X) / 2,
		Y: (p.Y + q.Y) / 2,
	}
}

// build the segments from the resolved contour points
func buildSegments(points []contourPoint) []Segment {
	var (
		firstOnCurveValid, firstOffCurveValid, lastOffCurveValid bool
		firstOnCurve, firstOffCurve, lastOffCurve                SegmentPoint
		out                                                      []Segment
	)

	for _, point := range points {
		p := point.SegmentPoint
		if !firstOnCurveValid {
			if point.isOnCurve {
				firstOnCurve = p
				firstOnCurveValid = true
				out = append(out, Segment{
					Op:   SegmentOpMoveTo,
					Args: [3]SegmentPoint{p},
				})
			} else if !firstOffCurveValid {
				firstOffCurve = p
				firstOffCurveValid = true

				if !point.isEndPoint {
					continue
				}
			} else {
				firstOnCurve = midPoint(firstOffCurve, p)
				firstOnCurveValid = true
				lastOffCurve = p
				lastOffCurveValid = true
				out = append(out, Segment{
					Op:   SegmentOpMoveTo,
					Args: [3]SegmentPoint{firstOnCurve},
				})
			}
		} else if !lastOffCurveValid {
			if !point.isOnCurve {
				lastOffCurve = p
				lastOffCurveValid = true

				if !point.isEndPoint {
					continue
				}
			} else {
				out = append(out, Segment{
					Op:   SegmentOpLineTo,
					Args: [3]SegmentPoint{p},
				})
			}
		} else {
			if !point.isOnCurve {
				out = append(out, Segment{
					Op: SegmentOpQuadTo,
					Args: [3]SegmentPoint{
						lastOffCurve,
						midPoint(lastOffCurve, p),
					},
				})
				lastOffCurve = p
				lastOffCurveValid = true
			} else {
				out = append(out, Segment{
					Op:   SegmentOpQuadTo,
					Args: [3]SegmentPoint{lastOffCurve, p},
				})
				lastOffCurveValid = false
			}
		}

		if point.isEndPoint {
			// closing the contour
			switch {
			case !firstOffCurveValid && !lastOffCurveValid:
				out = append(out, Segment{
					Op:   SegmentOpLineTo,
					Args: [3]SegmentPoint{firstOnCurve},
				})
			case !firstOffCurveValid && lastOffCurveValid:
				out = append(out, Segment{
					Op:   SegmentOpQuadTo,
					Args: [3]SegmentPoint{lastOffCurve, firstOnCurve},
				})
			case firstOffCurveValid && !lastOffCurveValid:
				out = append(out, Segment{
					Op:   SegmentOpQuadTo,
					Args: [3]SegmentPoint{firstOffCurve, firstOnCurve},
				})
			case firstOffCurveValid && lastOffCurveValid:
				out = append(out, Segment{
					Op: SegmentOpQuadTo,
					Args: [3]SegmentPoint{
						lastOffCurve,
						midPoint(lastOffCurve, firstOffCurve),
					},
				},
					Segment{
						Op:   SegmentOpQuadTo,
						Args: [3]SegmentPoint{firstOffCurve, firstOnCurve},
					},
				)
			}

			firstOnCurveValid = false
			firstOffCurveValid = false
			lastOffCurveValid = false
		}
	}

	return out
}
