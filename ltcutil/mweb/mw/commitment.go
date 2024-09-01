package mw

import (
	"encoding/binary"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

type Commitment [33]byte

func generatorH() *secp256k1.JacobianPoint {
	var generatorH = [64]byte{
		0x50, 0x92, 0x9b, 0x74, 0xc1, 0xa0, 0x49, 0x54,
		0xb7, 0x8b, 0x4b, 0x60, 0x35, 0xe9, 0x7a, 0x5e,
		0x07, 0x8a, 0x5a, 0x0f, 0x28, 0xec, 0x96, 0xd5,
		0x47, 0xbf, 0xee, 0x9a, 0xce, 0x80, 0x3a, 0xc0,
		0x31, 0xd3, 0xc6, 0x86, 0x39, 0x73, 0x92, 0x6e,
		0x04, 0x9e, 0x63, 0x7c, 0xb1, 0xb5, 0xf4, 0x0a,
		0x36, 0xda, 0xc2, 0x8a, 0xf1, 0x76, 0x69, 0x68,
		0xc3, 0x0c, 0x23, 0x13, 0xf3, 0xa3, 0x89, 0x04,
	}
	var H secp256k1.JacobianPoint
	H.X.SetByteSlice(generatorH[:32])
	H.Y.SetByteSlice(generatorH[32:])
	H.Z.SetInt(1)
	return &H
}

func NewCommitment(blind *BlindingFactor, value uint64) *Commitment {
	var v secp256k1.ModNScalar
	var b, r secp256k1.JacobianPoint
	v.SetByteSlice(binary.BigEndian.AppendUint64(nil, value))
	secp256k1.ScalarBaseMultNonConst(blind.scalar(), &b)
	secp256k1.ScalarMultNonConst(&v, generatorH(), &r)
	secp256k1.AddNonConst(&b, &r, &r)
	return toCommitment(&r)
}

func toCommitment(r *secp256k1.JacobianPoint) *Commitment {
	r.ToAffine()
	c := &Commitment{8}
	r.X.PutBytesUnchecked(c[1:])
	if !r.X.SquareRootVal(&r.Y) {
		c[0]++
	}
	return c
}

func SwitchCommit(blind *BlindingFactor, value uint64) *Commitment {
	return NewCommitment(BlindSwitch(blind, value), value)
}

func (c *Commitment) toJacobian() *secp256k1.JacobianPoint {
	var r secp256k1.JacobianPoint
	var t secp256k1.FieldVal
	if r.X.SetByteSlice(c[1:]) {
		panic("overflowed")
	}
	if !r.Y.SquareRootVal(t.SquareVal(&r.X).Mul(&r.X).AddInt(7)) {
		panic("invalid commitment")
	}
	if c[0]&1 > 0 {
		r.Y.Negate(1)
	}
	r.Z.SetInt(1)
	return &r
}

func (pk *PublicKey) Commitment() *Commitment {
	return toCommitment(pk.toJacobian())
}

func (c *Commitment) PubKey() *PublicKey {
	return toPubKey(c.toJacobian())
}

func (c *Commitment) Add(c2 *Commitment) *Commitment {
	r := c2.toJacobian()
	secp256k1.AddNonConst(c.toJacobian(), r, r)
	return toCommitment(r)
}

func (c *Commitment) Sub(c2 *Commitment) *Commitment {
	r := c2.toJacobian()
	r.Y.Negate(1)
	secp256k1.AddNonConst(c.toJacobian(), r, r)
	return toCommitment(r)
}
