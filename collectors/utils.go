package collectors

import (
	"code.cloudfoundry.org/cli/types"
)


func BoolToFloat(val *bool) float64 {
	if val == nil {
		return 0
	}
	if *val {
		return 1
	}
	return 0
}

func NullIntToFloat(val *types.NullInt) float64 {
	if val == nil {
		return -1
	}
	if !val.IsSet {
		return -1
	}
	return float64(val.Value)
}
