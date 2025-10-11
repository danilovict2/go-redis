package set

type GeoSetScore struct {
	Longitude float64
	Latitude  float64
}

func (this GeoSetScore) Less(other SetScore) bool {
	o, ok := other.(GeoSetScore)
	if !ok {
		return false
	}

	return this.Longitude < o.Longitude || this.Latitude < o.Latitude
}

func (this GeoSetScore) Equal(other SetScore) bool {
	o, ok := other.(GeoSetScore)
	if !ok {
		return false
	}

	return this.Longitude == o.Longitude && this.Latitude == o.Latitude
}
