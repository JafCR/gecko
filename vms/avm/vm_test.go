// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package avm

import (
	"bytes"
	"testing"

	"github.com/ava-labs/gecko/database/memdb"
	"github.com/ava-labs/gecko/ids"
	"github.com/ava-labs/gecko/snow"
	"github.com/ava-labs/gecko/snow/engine/common"
	"github.com/ava-labs/gecko/utils/crypto"
	"github.com/ava-labs/gecko/utils/formatting"
	"github.com/ava-labs/gecko/utils/hashing"
	"github.com/ava-labs/gecko/utils/units"
	"github.com/ava-labs/gecko/vms/components/codec"
	"github.com/ava-labs/gecko/vms/secp256k1fx"
)

var networkID uint32 = 43110
var chainID = ids.NewID([32]byte{5, 4, 3, 2, 1})

var keys []*crypto.PrivateKeySECP256K1R
var ctx *snow.Context
var asset = ids.NewID([32]byte{1, 2, 3})

func init() {
	ctx = snow.DefaultContextTest()
	ctx.NetworkID = networkID
	ctx.ChainID = chainID
	cb58 := formatting.CB58{}
	factory := crypto.FactorySECP256K1R{}

	for _, key := range []string{
		"24jUJ9vZexUM6expyMcT48LBx27k1m7xpraoV62oSQAHdziao5",
		"2MMvUMsxx6zsHSNXJdFD8yc5XkancvwyKPwpw4xUK3TCGDuNBY",
		"cxb7KpGWhDMALTjNNSJ7UQkkomPesyWAPUaWRGdyeBNzR6f35",
	} {
		ctx.Log.AssertNoError(cb58.FromString(key))
		pk, err := factory.ToPrivateKey(cb58.Bytes)
		ctx.Log.AssertNoError(err)
		keys = append(keys, pk.(*crypto.PrivateKeySECP256K1R))
	}
}

func GetFirstTxFromGenesisTest(genesisBytes []byte, t *testing.T) *Tx {
	c := codec.NewDefault()
	c.RegisterType(&BaseTx{})
	c.RegisterType(&CreateAssetTx{})
	c.RegisterType(&OperationTx{})
	c.RegisterType(&secp256k1fx.MintOutput{})
	c.RegisterType(&secp256k1fx.TransferOutput{})
	c.RegisterType(&secp256k1fx.MintInput{})
	c.RegisterType(&secp256k1fx.TransferInput{})
	c.RegisterType(&secp256k1fx.Credential{})

	genesis := Genesis{}
	if err := c.Unmarshal(genesisBytes, &genesis); err != nil {
		t.Fatal(err)
	}

	for _, genesisTx := range genesis.Txs {
		if len(genesisTx.Outs) != 0 {
			t.Fatal("genesis tx can't have non-new assets")
		}

		tx := Tx{
			UnsignedTx: &genesisTx.CreateAssetTx,
		}
		txBytes, err := c.Marshal(&tx)
		if err != nil {
			t.Fatal(err)
		}
		tx.Initialize(txBytes)

		return &tx
	}

	t.Fatal("genesis tx didn't have any txs")
	return nil
}

func BuildGenesisTest(t *testing.T) []byte {
	ss := StaticService{}

	addr0 := keys[0].PublicKey().Address()
	addr1 := keys[1].PublicKey().Address()
	addr2 := keys[2].PublicKey().Address()

	args := BuildGenesisArgs{GenesisData: map[string]AssetDefinition{
		"asset1": AssetDefinition{
			Name:   "myFixedCapAsset",
			Symbol: "MFCA",
			InitialState: map[string][]interface{}{
				"fixedCap": []interface{}{
					Holder{
						Amount:  100000,
						Address: addr0.String(),
					},
					Holder{
						Amount:  100000,
						Address: addr0.String(),
					},
					Holder{
						Amount:  50000,
						Address: addr0.String(),
					},
					Holder{
						Amount:  50000,
						Address: addr0.String(),
					},
				},
			},
		},
		"asset2": AssetDefinition{
			Name:   "myVarCapAsset",
			Symbol: "MVCA",
			InitialState: map[string][]interface{}{
				"variableCap": []interface{}{
					Owners{
						Threshold: 1,
						Minters: []string{
							addr0.String(),
							addr1.String(),
						},
					},
					Owners{
						Threshold: 2,
						Minters: []string{
							addr0.String(),
							addr1.String(),
							addr2.String(),
						},
					},
				},
			},
		},
		"asset3": AssetDefinition{
			Name: "myOtherVarCapAsset",
			InitialState: map[string][]interface{}{
				"variableCap": []interface{}{
					Owners{
						Threshold: 1,
						Minters: []string{
							addr0.String(),
						},
					},
				},
			},
		},
	}}
	reply := BuildGenesisReply{}
	err := ss.BuildGenesis(nil, &args, &reply)
	if err != nil {
		t.Fatal(err)
	}

	return reply.Bytes.Bytes
}

func GenesisVM(t *testing.T) *VM {
	genesisBytes := BuildGenesisTest(t)

	ctx.Lock.Lock()
	defer ctx.Lock.Unlock()

	vm := &VM{}
	err := vm.Initialize(
		ctx,
		memdb.New(),
		genesisBytes,
		make(chan common.Message, 1),
		[]*common.Fx{&common.Fx{
			ID: ids.Empty,
			Fx: &secp256k1fx.Fx{},
		}},
	)
	if err != nil {
		t.Fatal(err)
	}
	vm.batchTimeout = 0

	return vm
}

func TestTxSerialization(t *testing.T) {
	expected := []byte{
		// txID:
		0x00, 0x00, 0x00, 0x02,
		// networkID:
		0x00, 0x00, 0xa8, 0x66,
		// chainID:
		0x05, 0x04, 0x03, 0x02, 0x01, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		// number of outs:
		0x00, 0x00, 0x00, 0x03,
		// output[0]:
		// assetID:
		0x01, 0x02, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		// fxID:
		0x00, 0x00, 0x00, 0x04,
		// secp256k1 Transferable Output:
		// amount:
		0x00, 0x00, 0x12, 0x30, 0x9c, 0xe5, 0x40, 0x00,
		// locktime:
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		// threshold:
		0x00, 0x00, 0x00, 0x01,
		// number of addresses
		0x00, 0x00, 0x00, 0x01,
		// address[0]
		0xfc, 0xed, 0xa8, 0xf9, 0x0f, 0xcb, 0x5d, 0x30,
		0x61, 0x4b, 0x99, 0xd7, 0x9f, 0xc4, 0xba, 0xa2,
		0x93, 0x07, 0x76, 0x26,
		// output[1]:
		// assetID:
		0x01, 0x02, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		// fxID:
		0x00, 0x00, 0x00, 0x04,
		// secp256k1 Transferable Output:
		// amount:
		0x00, 0x00, 0x12, 0x30, 0x9c, 0xe5, 0x40, 0x00,
		// locktime:
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		// threshold:
		0x00, 0x00, 0x00, 0x01,
		// number of addresses:
		0x00, 0x00, 0x00, 0x01,
		// address[0]:
		0x6e, 0xad, 0x69, 0x3c, 0x17, 0xab, 0xb1, 0xbe,
		0x42, 0x2b, 0xb5, 0x0b, 0x30, 0xb9, 0x71, 0x1f,
		0xf9, 0x8d, 0x66, 0x7e,
		// output[2]:
		// assetID:
		0x01, 0x02, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		// fxID:
		0x00, 0x00, 0x00, 0x04,
		// secp256k1 Transferable Output:
		// amount:
		0x00, 0x00, 0x12, 0x30, 0x9c, 0xe5, 0x40, 0x00,
		// locktime:
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		// threshold:
		0x00, 0x00, 0x00, 0x01,
		// number of addresses:
		0x00, 0x00, 0x00, 0x01,
		// address[0]:
		0xf2, 0x42, 0x08, 0x46, 0x87, 0x6e, 0x69, 0xf4,
		0x73, 0xdd, 0xa2, 0x56, 0x17, 0x29, 0x67, 0xe9,
		0x92, 0xf0, 0xee, 0x31,
		// number of inputs:
		0x00, 0x00, 0x00, 0x00,
		// number of operations:
		0x00, 0x00, 0x00, 0x01,
		// operation[0]:
		// assetID:
		0x01, 0x02, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		// number of inputs:
		0x00, 0x00, 0x00, 0x00,
		// number of outputs:
		0x00, 0x00, 0x00, 0x01,
		// fxID:
		0x00, 0x00, 0x00, 0x03,
		// secp256k1 Mint Output:
		// threshold:
		0x00, 0x00, 0x00, 0x01,
		// number of addresses:
		0x00, 0x00, 0x00, 0x01,
		// address[0]:
		0xfc, 0xed, 0xa8, 0xf9, 0x0f, 0xcb, 0x5d, 0x30,
		0x61, 0x4b, 0x99, 0xd7, 0x9f, 0xc4, 0xba, 0xa2,
		0x93, 0x07, 0x76, 0x26,
		// number of credentials:
		0x00, 0x00, 0x00, 0x00,
	}

	unsignedTx := &OperationTx{
		BaseTx: BaseTx{
			NetID: networkID,
			BCID:  chainID,
		},
		Ops: []*Operation{
			&Operation{
				Asset: Asset{
					ID: asset,
				},
				Outs: []*OperableOutput{
					&OperableOutput{
						Out: &secp256k1fx.MintOutput{
							OutputOwners: secp256k1fx.OutputOwners{
								Threshold: 1,
								Addrs:     []ids.ShortID{keys[0].PublicKey().Address()},
							},
						},
					},
				},
			},
		},
	}
	tx := &Tx{UnsignedTx: unsignedTx}
	for _, key := range keys {
		addr := key.PublicKey().Address()

		unsignedTx.Outs = append(unsignedTx.Outs, &TransferableOutput{
			Asset: Asset{
				ID: asset,
			},
			Out: &secp256k1fx.TransferOutput{
				Amt: 20 * units.KiloAva,
				OutputOwners: secp256k1fx.OutputOwners{
					Threshold: 1,
					Addrs:     []ids.ShortID{addr},
				},
			},
		})
	}

	c := codec.NewDefault()
	c.RegisterType(&BaseTx{})
	c.RegisterType(&CreateAssetTx{})
	c.RegisterType(&OperationTx{})
	c.RegisterType(&secp256k1fx.MintOutput{})
	c.RegisterType(&secp256k1fx.TransferOutput{})
	c.RegisterType(&secp256k1fx.MintInput{})
	c.RegisterType(&secp256k1fx.TransferInput{})
	c.RegisterType(&secp256k1fx.Credential{})

	b, err := c.Marshal(tx)
	if err != nil {
		t.Fatal(err)
	}
	tx.Initialize(b)

	result := tx.Bytes()
	if !bytes.Equal(expected, result) {
		t.Fatalf("\nExpected: 0x%x\nResult:   0x%x", expected, result)
	}
}

func TestInvalidGenesis(t *testing.T) {
	ctx.Lock.Lock()
	defer ctx.Lock.Unlock()

	vm := &VM{}
	err := vm.Initialize(
		/*context=*/ ctx,
		/*db=*/ memdb.New(),
		/*genesisState=*/ nil,
		/*engineMessenger=*/ make(chan common.Message, 1),
		/*fxs=*/ nil,
	)
	if err == nil {
		t.Fatalf("Should have errored due to an invalid genesis")
	}
}

func TestInvalidFx(t *testing.T) {
	genesisBytes := BuildGenesisTest(t)

	ctx.Lock.Lock()
	defer ctx.Lock.Unlock()

	vm := &VM{}
	err := vm.Initialize(
		/*context=*/ ctx,
		/*db=*/ memdb.New(),
		/*genesisState=*/ genesisBytes,
		/*engineMessenger=*/ make(chan common.Message, 1),
		/*fxs=*/ []*common.Fx{
			nil,
		},
	)
	if err == nil {
		t.Fatalf("Should have errored due to an invalid interface")
	}
}

func TestFxInitializationFailure(t *testing.T) {
	genesisBytes := BuildGenesisTest(t)

	ctx.Lock.Lock()
	defer ctx.Lock.Unlock()

	vm := &VM{}
	err := vm.Initialize(
		/*context=*/ ctx,
		/*db=*/ memdb.New(),
		/*genesisState=*/ genesisBytes,
		/*engineMessenger=*/ make(chan common.Message, 1),
		/*fxs=*/ []*common.Fx{&common.Fx{
			ID: ids.Empty,
			Fx: &testFx{initialize: errUnknownFx},
		}},
	)
	if err == nil {
		t.Fatalf("Should have errored due to an invalid fx initialization")
	}
}

type testTxBytes struct{ unsignedBytes []byte }

func (tx *testTxBytes) UnsignedBytes() []byte { return tx.unsignedBytes }

func TestIssueTx(t *testing.T) {
	genesisBytes := BuildGenesisTest(t)

	issuer := make(chan common.Message, 1)

	ctx.Lock.Lock()
	vm := &VM{}
	err := vm.Initialize(
		ctx,
		memdb.New(),
		genesisBytes,
		issuer,
		[]*common.Fx{&common.Fx{
			ID: ids.Empty,
			Fx: &secp256k1fx.Fx{},
		}},
	)
	if err != nil {
		t.Fatal(err)
	}
	vm.batchTimeout = 0

	genesisTx := GetFirstTxFromGenesisTest(genesisBytes, t)

	newTx := &Tx{UnsignedTx: &OperationTx{BaseTx: BaseTx{
		NetID: networkID,
		BCID:  chainID,
		Ins: []*TransferableInput{
			&TransferableInput{
				UTXOID: UTXOID{
					TxID:        genesisTx.ID(),
					OutputIndex: 1,
				},
				Asset: Asset{
					ID: genesisTx.ID(),
				},
				In: &secp256k1fx.TransferInput{
					Amt: 50000,
					Input: secp256k1fx.Input{
						SigIndices: []uint32{
							0,
						},
					},
				},
			},
		},
	}}}

	unsignedBytes, err := vm.codec.Marshal(&newTx.UnsignedTx)
	if err != nil {
		t.Fatal(err)
	}

	key := keys[0]
	sig, err := key.Sign(unsignedBytes)
	if err != nil {
		t.Fatal(err)
	}
	fixedSig := [crypto.SECP256K1RSigLen]byte{}
	copy(fixedSig[:], sig)

	newTx.Creds = append(newTx.Creds, &Credential{
		Cred: &secp256k1fx.Credential{
			Sigs: [][crypto.SECP256K1RSigLen]byte{
				fixedSig,
			},
		},
	})

	b, err := vm.codec.Marshal(newTx)
	if err != nil {
		t.Fatal(err)
	}
	newTx.Initialize(b)

	txID, err := vm.IssueTx(newTx.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if !txID.Equals(newTx.ID()) {
		t.Fatalf("Issue Tx returned wrong TxID")
	}
	ctx.Lock.Unlock()

	msg := <-issuer
	if msg != common.PendingTxs {
		t.Fatalf("Wrong message")
	}

	if txs := vm.PendingTxs(); len(txs) != 1 {
		t.Fatalf("Should have returned %d tx(s)", 1)
	}
}

func TestGenesisGetUTXOs(t *testing.T) {
	genesisBytes := BuildGenesisTest(t)

	ctx.Lock.Lock()
	vm := &VM{}
	err := vm.Initialize(
		ctx,
		memdb.New(),
		genesisBytes,
		make(chan common.Message, 1),
		[]*common.Fx{&common.Fx{
			ID: ids.Empty,
			Fx: &secp256k1fx.Fx{},
		}},
	)
	if err != nil {
		t.Fatal(err)
	}
	vm.batchTimeout = 0

	shortAddr := keys[0].PublicKey().Address()
	addr := ids.NewID(hashing.ComputeHash256Array(shortAddr.Bytes()))

	addrs := ids.Set{}
	addrs.Add(addr)
	utxos, err := vm.GetUTXOs(addrs)
	if err != nil {
		t.Fatal(err)
	}
	vm.Shutdown()
	ctx.Lock.Unlock()

	if len(utxos) != 7 {
		t.Fatalf("Wrong number of utxos (%d) returned", len(utxos))
	}
}
