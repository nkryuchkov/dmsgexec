package dmsgexec

import (
	"context"
	"encoding/hex"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/skycoin/dmsg"
	"github.com/skycoin/dmsg/cipher"
	"github.com/skycoin/dmsg/disc"
	"github.com/skycoin/skycoin/src/util/logging"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/nettest"
)

func TestNewServer(t *testing.T) {
	// prepare PKs
	aPK, aSK, err := cipher.GenerateDeterministicKeyPair([]byte("a seed"))
	require.NoError(t, err)
	bPK, bSK, err := cipher.GenerateDeterministicKeyPair([]byte("b seed"))
	require.NoError(t, err)

	// prepare auth file
	authF, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)
	authFName := authF.Name()
	defer func() { require.NoError(t, os.Remove(authFName)) }()
	require.NoError(t, authF.Close())
	auth := NewJsonFileWhiteList(authFName)
	require.NoError(t, auth.Add(aPK, bPK))

	t.Run("Whitelist_Get", func(t *testing.T) {
		for _, pk := range []cipher.PubKey{aPK, bPK} {
			ok, err := auth.Get(pk)
			require.NoError(t, err)
			require.True(t, ok)
		}
	})

	// prepare dmsg env
	dmsgD := disc.NewMock()
	sPK, sSK, err := cipher.GenerateDeterministicKeyPair([]byte("dmsg server seed"))
	require.NoError(t, err)
	sL, err := nettest.NewLocalListener("tcp")
	require.NoError(t, err)
	defer func() { _ = sL.Close() }() //nolint:errcheck
	dmsgS, err := dmsg.NewServer(sPK, sSK, "", sL, dmsgD)
	require.NoError(t, err)
	go func() { _ = dmsgS.Serve() }() //nolint:errcheck

	// test exec

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	aConf := ServerConfig{
		PubKey:   aPK,
		SecKey:   aSK,
		DmsgPort: DefaultDmsgPort,
		CLINet:   "unix",
		CLIAddr:  filepath.Join(os.TempDir(), hex.EncodeToString(cipher.RandByte(10))),
	}
	aSrv := Server{
		log:   logging.MustGetLogger("a_srv"),
		conf:  aConf,
		auth:  auth,
		dmsgC: dmsg.NewClient(aPK, aSK, dmsgD, dmsg.SetLogger(logging.MustGetLogger("a_dmsg"))),
	}
	go func() { _ = aSrv.Serve(ctx) }() //nolint:errcheck

	bConf := ServerConfig{
		PubKey:   bPK,
		SecKey:   bSK,
		DmsgPort: DefaultDmsgPort,
		CLINet:   "unix",
		CLIAddr:  filepath.Join(os.TempDir(), hex.EncodeToString(cipher.RandByte(10))),
	}
	bSrv := Server{
		log:   logging.MustGetLogger("b_srv"),
		conf:  bConf,
		auth:  auth,
		dmsgC: dmsg.NewClient(bPK, bSK, dmsgD, dmsg.SetLogger(logging.MustGetLogger("b_dmsg"))),
	}
	go func() { _ = bSrv.Serve(ctx) }() //nolint:errcheck

	time.Sleep(time.Second * 5)

	conn, err := net.Dial(aConf.CLINet, aConf.CLIAddr)
	require.NoError(t, err)

	b, err := Exec(conn, Command{
		DstPK:   bPK,
		DstPort: DefaultDmsgPort,
		Name:    "echo",
		Args:    []string{"hello world"},
	})
	require.NoError(t, err)
	require.Equal(t, "hello world", strings.TrimSpace(string(b)))
}
