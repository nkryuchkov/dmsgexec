package dmsgexec

import (
	"context"
	"io"
	"net/rpc"
	"os/exec"

	"github.com/skycoin/dmsg"
	"github.com/skycoin/dmsg/cipher"
	"github.com/skycoin/skycoin/src/util/logging"
)

type Command struct {
	DstPK   cipher.PubKey `json:"dst_pk"`
	DstPort uint16        `json:"dst_port"`
	Name    string        `json:"name"`
	Args    []string      `json:"args"`
}

type DmsgGateway struct {
	ctx context.Context
}

func NewDmsgGateway(ctx context.Context) *DmsgGateway {
	return &DmsgGateway{ctx: ctx}
}

func (g *DmsgGateway) Exec(in *Command, out *[]byte) error {
	cmd := exec.CommandContext(g.ctx, in.Name, in.Args...)
	b, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	*out = b
	return nil
}

type CLIGateway struct {
	ctx   context.Context
	log  *logging.Logger
	auth  Whitelist
	dmsgC *dmsg.Client
}

func NewCLIGateway(ctx context.Context, log *logging.Logger, auth Whitelist, dmsgC *dmsg.Client) *CLIGateway {
	return &CLIGateway{
		ctx:   ctx,
		log:   log,
		auth:  auth,
		dmsgC: dmsgC,
	}
}

func (g *CLIGateway) Exec(in *Command, out *[]byte) error {
	g.log.WithField("request", in).Info("Attempting exec.")

	conn, err := g.dmsgC.Dial(g.ctx, in.DstPK, in.DstPort)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	call := rpc.NewClient(conn).Go("DmsgGateway.Exec", in, out, nil)
	select {
	case <-g.ctx.Done():
		return g.ctx.Err()
	case <-call.Done:
		return call.Error
	}
}

func (g *CLIGateway) Status(_ *struct{}, out *bool) error {
	*out = true
	return nil
}

func (g *CLIGateway) Whitelist(_ *struct{}, out *[]cipher.PubKey) error {
	pks, err := g.auth.All()
	if err != nil {
		return err
	}
	*out = make([]cipher.PubKey, 0, len(pks))
	for pk, ok := range pks {
		if ok {
			*out = append(*out, pk)
		}
	}
	return nil
}

func (g *CLIGateway) WhitelistAdd(in *[]cipher.PubKey, _ *struct{}) error {
	return g.auth.Add(*in...)
}

func (g *CLIGateway) WhitelistRemove(in *[]cipher.PubKey, _ *struct{}) error {
	return g.auth.Remove(*in...)
}

func Exec(conn io.ReadWriteCloser, cmd Command) ([]byte, error) {
	var out []byte
	err := rpc.NewClient(conn).Call("CLIGateway.Exec", &cmd, &out)
	return out, err
}

/*
	RPC client side operations.
*/

// Used for RPC calls
var empty struct{}

func Status(conn io.ReadWriteCloser) (bool, error) {
	var out bool
	err := rpc.NewClient(conn).Call("CLIGateway.Status", &empty, &out)
	return out, err
}

func ViewWhitelist(conn io.ReadWriteCloser) ([]cipher.PubKey, error) {
	var pks []cipher.PubKey
	err := rpc.NewClient(conn).Call("CLIGateway.Whitelist", &empty, &pks)
	return pks, err
}

func WhitelistAdd(conn io.ReadWriteCloser, pks ...cipher.PubKey) error {
	return rpc.NewClient(conn).Call("CLIGateway.WhitelistAdd", &pks, &empty)
}

func WhitelistRemove(conn io.ReadWriteCloser, pks ...cipher.PubKey) error {
	return rpc.NewClient(conn).Call("CLIGateway.WhitelistRemove", &pks, &empty)
}
