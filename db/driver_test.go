package db

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"io"
	"testing"

	"github.com/eliothedeman/heath/block"
	"github.com/eliothedeman/randutil"
	"github.com/spf13/afero"
)

var testFS = afero.NewMemMapFs()

func newTestFile(name string) (io.ReadWriteSeeker, func() error) {
	f, _ := testFS.Create(name)
	return f, func() error {
		return f.Close()
	}
}

func newKey() *ecdsa.PrivateKey {
	k, _ := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	return k
}
func newTestBlock(s *block.Signature, priv *ecdsa.PrivateKey) *block.Block {
	b, _ := block.NewBlock(s, priv, randutil.Bytes(randutil.Int()%1000))
	return b
}

func TestDriversWrite(t *testing.T) {
	for name, df := range drivers {
		t.Run(fmt.Sprintf("Driver:%s", name), func(t *testing.T) {
			f, close := newTestFile(name)
			d := df(f, close)
			key := newKey()

			b := newTestBlock(nil, key)
			err := d.Write(b)
			if err != nil {
				t.Error(err)
			}

			x, xErr := d.GetBlockByContentHash(b.Signature.GetContentHash())
			if xErr != nil {
				t.Error(xErr)
			}

			if !bytes.Equal(x.GetPayload(), b.GetPayload()) {
				t.Error(*x, *b)
			}

			close()
		})
	}
}

func TestDriversRead(t *testing.T) {
	for name, df := range drivers {
		t.Run(fmt.Sprintf("Driver:%s", name), func(t *testing.T) {
			f, close := newTestFile(name)
			d := df(f, close)
			key := newKey()

			var b *block.Block
			b = newTestBlock(nil, key)
			for i := 0; i < 100; i++ {
				b = newTestBlock(b.GetSignature(), key)
				err := d.Write(b)
				if err != nil {
					t.Error(err)
				}
			}

			var count = 0
			out, err := d.StreamBlocks(context.Background())
			for b = range out {
				count++
			}
			if count != 100 {
				t.Fail()
			}

			verr := <-err

			if verr != nil {
				t.Error(verr)
			}

			close()
		})
	}

}
