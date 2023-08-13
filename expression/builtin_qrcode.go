package expression

import (
	"bytes"
	"fmt"

	"github.com/pingcap/tidb/sessionctx"
	"github.com/pingcap/tidb/types"
	"github.com/pingcap/tidb/util/chunk"

	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

var (
	_ functionClass = &qrcodeFunctionClass{}
)

var (
	_ builtinFunc = &builtinQrcodeSig{}
)

type qrBufWriter struct {
	*bytes.Buffer
}

func (b *qrBufWriter) Close() error {
	return nil
}

type qrcodeFunctionClass struct {
	baseFunctionClass
}

func (c *qrcodeFunctionClass) getFunction(ctx sessionctx.Context, args []Expression) (builtinFunc, error) {
	if err := c.verifyArgs(args); err != nil {
		return nil, err
	}

	bf, err := newBaseBuiltinFuncWithTp(ctx, c.funcName, args, types.ETString, types.ETString)
	if err != nil {
		return nil, err
	}

	types.SetBinChsClnFlag(bf.tp)
	sig := &builtinQrcodeSig{bf}
	return sig, nil
}

type builtinQrcodeSig struct {
	baseBuiltinFunc
}

func (b *builtinQrcodeSig) Clone() builtinFunc {
	newSig := &builtinQrcodeSig{}
	newSig.cloneFrom(&b.baseBuiltinFunc)
	return newSig
}

func (b *builtinQrcodeSig) evalString(row chunk.Row) (string, bool, error) {
	d, isNull, err := b.args[0].EvalString(b.ctx, row)
	if isNull || err != nil {
		return d, isNull, err
	}
	qrc, err := qrcode.New(d)
	if err != nil {
		fmt.Printf("could not generate QRCode: %v", err)
		return "", isNull, err
	}

	var bb bytes.Buffer
	qrbw := &qrBufWriter{&bb}
	qrWriter := standard.NewWithWriter(qrbw)

	if err = qrc.Save(qrWriter); err != nil {
		fmt.Printf("could not save image: %v", err)
		return "", isNull, err
	}

	return qrbw.String(), isNull, nil
}
