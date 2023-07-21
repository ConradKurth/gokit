package hashfunc

import (
	"hash/fnv"
	"strconv"

	"github.com/ConradKurth/gokit/config"
	hashids "github.com/speps/go-hashids/v2"
)

// Hash takes a string and returns a hash uint32 representation
func Hash(s string) (uint32, error) {
	h := fnv.New32a()
	if _, err := h.Write([]byte(s)); err != nil {
		return 0, err
	}
	return h.Sum32(), nil
}

// Hasher will hash a string id
type Hasher interface {
	HashID(s string) (string, error)
	HashString(s string) (string, error)
}

// IDHasher contains our config from the hash secret
type IDHasher struct {
	config *config.Config
}

// New will create a new hasher
func New(c *config.Config) *IDHasher {
	return &IDHasher{
		config: c,
	}
}

// HashID will hash the given ID with a specific alphabete
func (h *IDHasher) HashID(s string) (string, error) {
	hd := hashids.NewData()
	hd.MinLength = 5
	hd.Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	hd.Salt = h.config.GetString("hash.salt")
	hasher, err := hashids.NewWithData(hd)
	if err != nil {
		return "", err
	}

	return hasher.EncodeHex(s)
}

// HashString will hash the given string with a specific alphabete
func (h *IDHasher) HashString(s string) (string, error) {

	u, err := Hash(s)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(int(u)), nil
}
