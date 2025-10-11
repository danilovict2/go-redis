package set

import "fmt"

const (
	MinLatitude  float64 = -85.05112878
	MaxLatitude  float64 = 85.05112878
	MinLongitude float64 = -180
	MaxLongitude float64 = 180

	LatitudeRange  float64 = MaxLatitude - MinLatitude
	LongitudeRange float64 = MaxLongitude - MinLongitude
)

type GeoSetScore struct {
	Longitude float64
	Latitude  float64
	score     int // Used for comparison
}

func (s GeoSetScore) String() string {
	return fmt.Sprint(s.score)
}

func NewGeoSetScore(long, lat float64) GeoSetScore {
	return GeoSetScore{
		Longitude: long,
		Latitude:  lat,
		score:     generateScoreFrom(long, lat),
	}
}

func generateScoreFrom(long, lat float64) int {
	normalizedLong := int(1 << 26 * (long - MinLongitude) / LongitudeRange)
	normalizedLat := int(1 << 26 * (lat - MinLatitude) / LatitudeRange)
	return interleave(normalizedLat, normalizedLong)
}

func interleave(normalizedLat, normalizedLong int) int {
	x := spreadInt32ToInt64(normalizedLat)
	y := spreadInt32ToInt64(normalizedLong)
	yShifted := y << 1

	return x | yShifted
}

func spreadInt32ToInt64(x int) int {
	x = x & 0xFFFFFFFF

	x = (x | (x << 16)) & 0x0000FFFF0000FFFF
	x = (x | (x << 8)) & 0x00FF00FF00FF00FF
	x = (x | (x << 4)) & 0x0F0F0F0F0F0F0F0F
	x = (x | (x << 2)) & 0x3333333333333333
	x = (x | (x << 1)) & 0x5555555555555555

	return x
}

func (this GeoSetScore) Less(other SetScore) bool {
	o, ok := other.(GeoSetScore)
	if !ok {
		return false
	}

	return this.score < o.score
}

func (this GeoSetScore) Equal(other SetScore) bool {
	o, ok := other.(GeoSetScore)
	if !ok {
		return false
	}

	return this.score == o.score
}
