package geohash

const (
	MinLatitude  float64 = -85.05112878
	MaxLatitude  float64 = 85.05112878
	MinLongitude float64 = -180
	MaxLongitude float64 = 180

	LatitudeRange  float64 = MaxLatitude - MinLatitude
	LongitudeRange float64 = MaxLongitude - MinLongitude
)

type DecodedGeoScore struct {
	Long float64
	Lat  float64
}

func EncodeGeoScore(long, lat float64) int {
	return generateScoreFrom(long, lat)
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

func DecodeGeoScore(score int) DecodedGeoScore {
	y := score >> 1
	x := score

	gridLatNumber := compactInt64ToInt32(x)
	gridLongNumber := compactInt64ToInt32(y)

	grid_latitude_min := MinLatitude + LatitudeRange*(float64(gridLatNumber)/(1<<26))
	grid_latitude_max := MinLatitude + LatitudeRange*((float64(gridLatNumber)+1)/(1<<26))
	grid_longitude_min := MinLongitude + LongitudeRange*(float64(gridLongNumber)/(1<<26))
	grid_longitude_max := MinLongitude + LongitudeRange*(float64(gridLongNumber+1)/(1<<26))

	lat := (grid_latitude_min + grid_latitude_max) / 2
	long := (grid_longitude_min + grid_longitude_max) / 2

	return DecodedGeoScore{Long: long, Lat: lat}
}

func compactInt64ToInt32(x int) int {
	x = x & 0x5555555555555555
	x = (x | (x >> 1)) & 0x3333333333333333
	x = (x | (x >> 2)) & 0x0F0F0F0F0F0F0F0F
	x = (x | (x >> 4)) & 0x00FF00FF00FF00FF
	x = (x | (x >> 8)) & 0x0000FFFF0000FFFF
	x = (x | (x >> 16)) & 0x00000000FFFFFFFF

	return x
}
