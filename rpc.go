package dmsgexec

import (
	"context"
	"io"
	"net/rpc"
	"os/exec"

	"github.com/skycoin/dmsg"
	"github.com/skycoin/dmsg/cipher"
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
	auth  Whitelist
	dmsgC *dmsg.Client
}

func NewCLIGateway(ctx context.Context, auth Whitelist, dmsgC *dmsg.Client) *CLIGateway {
	return &CLIGateway{
		ctx:   ctx,
		auth:  auth,
		dmsgC: dmsgC,
	}
}

func (g *CLIGateway) Exec(in *Command, out *[]byte) error {
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

func Exec(conn io.ReadWriteCloser, cmd Command) ([]byte, error) {
	var out []byte
	err := rpc.NewClient(conn).Call("CLIGateway.Exec", &cmd, &out)
	return out, err
}

func Status(conn io.ReadWriteCloser) (bool, error) {
	var out bool
	err := rpc.NewClient(conn).Call("CLIGateway.Status", &struct{}{}, &out)
	return out, err
}
