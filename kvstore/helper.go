package kvstore

import (
	"fmt"
	"strconv"

	"github.com/garyburd/redigo/redis"
)

// Ints is a helper that converts a multi-bulk command reply to a []int.
// If err is not equal to nil, then Ints returns nil, err.  If one if the
// multi-bulk items is not a bulk value or nil, then Ints returns an error.
func Ints(reply interface{}, err error) ([]int, error) {
	if err != nil {
		return nil, err
	}
	switch reply := reply.(type) {
	case []interface{}:
		result := make([]int, len(reply))
		for i := range reply {
			if reply[i] == nil {
				continue
			}
			p, ok := reply[i].([]byte)
			if !ok {
				return nil, fmt.Errorf("kvstore: unexpected element type for Ints, got type %T", reply[i])
			}

			n, err := strconv.ParseInt(string(p), 10, 0)

			if err != nil {
				return nil, err
			}

			result[i] = int(n)
		}
		return result, nil
	case nil:
		return nil, redis.ErrNil
	case redis.Error:
		return nil, reply
	}
	return nil, fmt.Errorf("kvstore: unexpected type for Ints, got type %T", reply)
}
