// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package spdagvm

import (
	"bytes"
	"testing"

	"github.com/ava-labs/gecko/ids"
)

func TestOutputPayment(t *testing.T) {
	addr0 := ids.NewShortID([20]byte{
		0x51, 0x02, 0x5c, 0x61, 0xfb, 0xcf, 0xc0, 0x78,
		0xf6, 0x93, 0x34, 0xf8, 0x34, 0xbe, 0x6d, 0xd2,
		0x6d, 0x55, 0xa9, 0x55,
	})
	addr1 := ids.NewShortID([20]byte{
		0xc3, 0x34, 0x41, 0x28, 0xe0, 0x60, 0x12, 0x8e,
		0xde, 0x35, 0x23, 0xa2, 0x4a, 0x46, 0x1c, 0x89,
		0x43, 0xab, 0x08, 0x59,
	})

	b := Builder{
		NetworkID: 0,
		ChainID:  ids.Empty,
	}
	output := b.NewOutputPayment(
		/*amount=*/ 12345,
		/*locktime=*/ 54321,
		/*threshold=*/ 1,
		/*addresses=*/ []ids.ShortID{addr0, addr1},
	)

	c := Codec{}
	outputBytes, err := c.MarshalOutput(output)
	if err != nil {
		t.Fatal(err)
	}

	expected := []byte{
		// output type
		0x00, 0x00, 0x00, 0x00,
		// amount:
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x30, 0x39,
		// locktime:
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xd4, 0x31,
		// threshold:
		0x00, 0x00, 0x00, 0x01,
		// number of addresses:
		0x00, 0x00, 0x00, 0x02,
		// addr0:
		0x51, 0x02, 0x5c, 0x61, 0xfb, 0xcf, 0xc0, 0x78,
		0xf6, 0x93, 0x34, 0xf8, 0x34, 0xbe, 0x6d, 0xd2,
		0x6d, 0x55, 0xa9, 0x55,
		// addr1:
		0xc3, 0x34, 0x41, 0x28, 0xe0, 0x60, 0x12, 0x8e,
		0xde, 0x35, 0x23, 0xa2, 0x4a, 0x46, 0x1c, 0x89,
		0x43, 0xab, 0x08, 0x59,
	}
	if !bytes.Equal(outputBytes, expected) {
		t.Fatalf("Codec.MarshalOutput returned:\n0x%x\nExpected:\n0x%x", outputBytes, expected)
	}
}

func TestOutputTakeOrLeave(t *testing.T) {
	addr0 := ids.NewShortID([20]byte{
		0x51, 0x02, 0x5c, 0x61, 0xfb, 0xcf, 0xc0, 0x78,
		0xf6, 0x93, 0x34, 0xf8, 0x34, 0xbe, 0x6d, 0xd2,
		0x6d, 0x55, 0xa9, 0x55,
	})
	addr1 := ids.NewShortID([20]byte{
		0xc3, 0x34, 0x41, 0x28, 0xe0, 0x60, 0x12, 0x8e,
		0xde, 0x35, 0x23, 0xa2, 0x4a, 0x46, 0x1c, 0x89,
		0x43, 0xab, 0x08, 0x59,
	})

	b := Builder{
		NetworkID: 0,
		ChainID:  ids.Empty,
	}
	output := b.NewOutputTakeOrLeave(
		/*amount=*/ 12345,
		/*locktime1=*/ 54321,
		/*threshold1=*/ 1,
		/*addresses1=*/ []ids.ShortID{addr0},
		/*locktime2=*/ 56789,
		/*threshold2=*/ 1,
		/*addresses2=*/ []ids.ShortID{addr1},
	)

	c := Codec{}
	outputBytes, err := c.MarshalOutput(output)
	if err != nil {
		t.Fatal(err)
	}

	expected := []byte{
		// output type
		0x00, 0x00, 0x00, 0x01,
		// amount:
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x30, 0x39,
		// locktime1:
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xd4, 0x31,
		// threshold1:
		0x00, 0x00, 0x00, 0x01,
		// number of addresses1:
		0x00, 0x00, 0x00, 0x01,
		// addr0:
		0x51, 0x02, 0x5c, 0x61, 0xfb, 0xcf, 0xc0, 0x78,
		0xf6, 0x93, 0x34, 0xf8, 0x34, 0xbe, 0x6d, 0xd2,
		0x6d, 0x55, 0xa9, 0x55,
		// locktime2:
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xdd, 0xd5,
		// threshold2:
		0x00, 0x00, 0x00, 0x01,
		// number of addresses2:
		0x00, 0x00, 0x00, 0x01,
		// addr1:
		0xc3, 0x34, 0x41, 0x28, 0xe0, 0x60, 0x12, 0x8e,
		0xde, 0x35, 0x23, 0xa2, 0x4a, 0x46, 0x1c, 0x89,
		0x43, 0xab, 0x08, 0x59,
	}
	if !bytes.Equal(outputBytes, expected) {
		t.Fatalf("Codec.MarshalOutput returned:\n0x%x\nExpected:\n0x%x", outputBytes, expected)
	}
}