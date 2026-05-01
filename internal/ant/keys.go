package ant

import (
	"fmt"

	sqids "github.com/sqids/sqids-go"
)

// idAlphabet is the alphabet used for the sqid suffix. Visually confusable
// characters (0/O, 1/I/l) are omitted so ids are easier to read aloud and
// type. 56 characters total — comfortably above sqids' minimum.
const idAlphabet = "23456789abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"

// idMinLength keeps short ids legible. sqids lengthens automatically as the
// integer grows, so this only affects the bottom end of the range.
const idMinLength = 5

var idEncoder *sqids.Sqids

func init() {
	s, err := sqids.New(sqids.Options{
		Alphabet:  idAlphabet,
		MinLength: idMinLength,
	})
	if err != nil {
		panic(fmt.Sprintf("ant: invalid sqids configuration: %v", err))
	}
	idEncoder = s
}

// PublicID generates the public identifier "<prefix>-<sqid>" for the given
// integer primary key. Errors on empty prefix or negative id.
func PublicID(prefix string, id int64) (string, error) {
	if prefix == "" {
		return "", fmt.Errorf("PublicID: empty prefix")
	}
	if id < 0 {
		return "", fmt.Errorf("PublicID: negative id %d", id)
	}
	encoded, err := idEncoder.Encode([]uint64{uint64(id)})
	if err != nil {
		return "", fmt.Errorf("PublicID: sqid encode: %w", err)
	}
	return prefix + "-" + encoded, nil
}
