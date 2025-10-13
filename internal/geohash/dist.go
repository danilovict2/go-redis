package geohash

import "math"

const rearth = 6372797.560856 // m

func haversine(θ float64) float64 {
	return .5 * (1 - math.Cos(θ))
}

type pos struct {
	φ float64 // latitude, radians
	ψ float64 // longitude, radians
}

func DegPos(lat, lon float64) pos {
	return pos{lat * math.Pi / 180, lon * math.Pi / 180}
}

func Hsdist(p1, p2 pos) float64 {
	return 2 * rearth * math.Asin(math.Sqrt(haversine(p2.φ-p1.φ)+
		math.Cos(p1.φ)*math.Cos(p2.φ)*haversine(p2.ψ-p1.ψ)))
}
