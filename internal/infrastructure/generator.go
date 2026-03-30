package infrastructure

import (
	"fmt"
	"math/rand/v2"
)

const DefaultKeyLen = 16

type generator struct {
	chacha  *rand.ChaCha8
	byteLen int
}

type GeneratorOption func(*generator)

func NewGenerator(seed [32]byte, opts ...GeneratorOption) *generator {
	g := &generator{
		chacha:  rand.NewChaCha8(seed),
		byteLen: DefaultKeyLen,
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

func WithByteLen(byteLen int) GeneratorOption {
	return func(g *generator) {
		g.byteLen = byteLen
	}
}

func (gen *generator) Generate() string {
	key := make([]byte, gen.byteLen/8)
	gen.chacha.Read(key)
	return fmt.Sprintf("%X", key)
}
