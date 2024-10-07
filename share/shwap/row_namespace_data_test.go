package shwap_test

import (
	"bytes"
	"context"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	gosquare "github.com/celestiaorg/go-square/v2/share"

	"github.com/celestiaorg/celestia-node/share"
	"github.com/celestiaorg/celestia-node/share/eds"
	"github.com/celestiaorg/celestia-node/share/eds/edstest"
	"github.com/celestiaorg/celestia-node/share/shwap"
)

func TestNamespacedRowFromShares(t *testing.T) {
	const odsSize = 8

	minNamespace, err := gosquare.NewV0Namespace(slices.Concat(bytes.Repeat([]byte{0}, 8), []byte{1, 0}))
	require.NoError(t, err)
	err = gosquare.ValidateForData(minNamespace)
	require.NoError(t, err)

	for namespacedAmount := 1; namespacedAmount < odsSize; namespacedAmount++ {
		shares := gosquare.RandSharesWithNamespace(minNamespace, namespacedAmount, odsSize)
		parity, err := share.DefaultRSMT2DCodec().Encode(gosquare.ToBytes(shares))
		require.NoError(t, err)

		paritySh, err := gosquare.FromBytes(parity)
		require.NoError(t, err)
		extended := slices.Concat(shares, paritySh)

		nr, err := shwap.RowNamespaceDataFromShares(extended, minNamespace, 0)
		require.NoError(t, err)
		require.Equal(t, namespacedAmount, len(nr.Shares))
	}
}

// func TestNamespacedRowFromSharesNonIncluded(t *testing.T) {
// //TODO: this will fail until absence proof support is added
//	t.Skip()
//
//	const odsSize = 8
//	//Test absent namespace
//	shares := gosquare.RandShares( odsSize)
//	absentNs, err := share.GetNamespace(shares[0]).AddInt(1)
//	require.NoError(t, err)
//
//	parity, err := share.DefaultRSMT2DCodec().Encode(shares)
//	require.NoError(t, err)
//	extended := slices.Concat(shares, parity)
//
//	shrs, err := gosquare.FromBytes(extended)
//	require.NoError(t, err)
//
//	nr, err := shwap.RowNamespaceDataFromShares(shrs, absentNs, 0)
//	require.NoError(t, err)
//	require.Len(t, nr.Shares, 0)
//	require.True(t, nr.Proof.IsOfAbsence())
//}

func TestValidateNamespacedRow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	const odsSize = 8
	sharesAmount := odsSize * odsSize
	namespace := gosquare.RandomNamespace()
	for amount := 1; amount < sharesAmount; amount++ {
		randEDS, root := edstest.RandEDSWithNamespace(t, namespace, amount, odsSize)
		rsmt2d := &eds.Rsmt2D{ExtendedDataSquare: randEDS}
		nd, err := eds.NamespaceData(ctx, rsmt2d, namespace)
		require.NoError(t, err)
		require.True(t, len(nd) > 0)

		rowIdxs := share.RowsWithNamespace(root, namespace)
		require.Len(t, nd, len(rowIdxs))

		for i, rowIdx := range rowIdxs {
			err = nd[i].Verify(root, namespace, rowIdx)
			require.NoError(t, err)
		}
	}
}

func TestNamespacedRowProtoEncoding(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	const odsSize = 8
	namespace := gosquare.RandomNamespace()
	randEDS, _ := edstest.RandEDSWithNamespace(t, namespace, odsSize, odsSize)
	rsmt2d := &eds.Rsmt2D{ExtendedDataSquare: randEDS}
	nd, err := eds.NamespaceData(ctx, rsmt2d, namespace)
	require.NoError(t, err)
	require.True(t, len(nd) > 0)

	expected := nd[0]
	pb := expected.ToProto()
	ndOut, err := shwap.RowNamespaceDataFromProto(pb)
	require.NoError(t, err)
	require.Equal(t, expected, ndOut)
}