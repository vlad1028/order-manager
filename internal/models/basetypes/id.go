package basetypes

import "strconv"

type ID uint64

func (id *ID) String() string {
	return strconv.FormatUint(uint64(*id), 10)
}
