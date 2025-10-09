package set

type RegularSetScore float64

func (this RegularSetScore) Less(other SetScore) bool {
	o, ok := other.(RegularSetScore)
	if !ok {
		return false
	}

	return float64(this) < float64(o)
}

func (this RegularSetScore) Equal(other SetScore) bool {
	o, ok := other.(RegularSetScore)
	if !ok {
		return false
	}

	return float64(this) == float64(o)

}
