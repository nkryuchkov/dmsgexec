package dmsgexec

import (
	"context"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"sync"

	"github.com/skycoin/dmsg"
	"github.com/skycoin/dmsg/cipher"
	"github.com/skycoin/dmsg/disc"
	"github.com/skycoin/skycoin/src/util/logging"
)

const (
	DefaultDmsgDisc = "http://messaging.discovery.skywire.skycoin.com"
	DefaultDmsgPort = uint16(222)
	DefaultCLINet   = "unix"
	DefaultCLIAddr  = "dmsgexec.sock"
)

type ServerConfig struct {
	PubKey   cipher.PubKey `json:"public_key"`
	SecKey   cipher.SecKey `json:"secret_key"`
	DmsgDisc string        `json:"dmsg_discovery"` // address for dmsg discovery to use
	DmsgPort uint16        `json:"dmsg_port"`      // local listening port for dmsgexec commands
	CLINet   string        `json:"cli_net"`        // local listening network for cli
	CLIAddr  string        `json:"cli_addr"`       // local listening address for cli
}

func DefaultServerConfig(pk cipher.PubKey, sk cipher.SecKey) ServerConfig {
	return ServerConfig{
		PubKey: pk,
		SecKey: sk,
		DmsgDisc: DefaultDmsgDisc,
		DmsgPort: DefaultDmsgPort,
		CLINet: DefaultCLINet,
		CLIAddr: DefaultCLIAddr,
	}
}

type Server struct {
	log *logging.Logger

	conf  ServerConfig
	auth  Whitelist
	dmsgC *dmsg.Client
}

func NewServer(auth Whitelist, conf ServerConfig) *Server {
	return &Server{
		log:   logging.MustGetLogger("dmsgexec_server"),
		conf:  conf,
		auth:  auth,
		dmsgC: dmsg.NewClient(conf.PubKey, conf.SecKey, disc.NewHTTP(conf.DmsgDisc)),
	}
}

func (s *Server) Serve(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := s.dmsgC.InitiateServerConnections(ctx, 1); err != nil {
		return err
	}
	dmsgL, err := s.dmsgC.Listen(s.conf.DmsgPort)
	if err != nil {
		return fmt.Errorf("failed to create dmsg listener: %v", err)
	}
	go func() {
		<-ctx.Done()
		dmsgL.Close() //nolint:errcheck
	}()

	// TODO: pid file.
	if s.conf.CLINet == "unix" {
		os.Remove(s.conf.CLIAddr) //nolint:errcheck
	}
	cliL, err := net.Listen(s.conf.CLINet, s.conf.CLIAddr)
	if err != nil {
		return fmt.Errorf("failed to create cli listener: %v", err)
	}
	go func() {
		<-ctx.Done()
		cliL.Close() //nolint:errcheck
	}()

	dmsgS := rpc.NewServer()
	if err := dmsgS.Register(NewDmsgGateway(ctx)); err != nil {
		return err
	}
	cliS := rpc.NewServer()
	if err := cliS.Register(NewCLIGateway(ctx, s.auth, s.dmsgC)); err != nil {
		return err
	}
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		defer wg.Done()
		for {
			conn, err := dmsgL.AcceptTransport()
			if err != nil {
				return
			}
			ok, err := s.auth.Get(conn.RemotePK())
			if err != nil {
				return
			}
			if !ok {
				conn.Close() //nolint:errcheck
				continue
			}
			go dmsgS.ServeConn(conn)
		}
	}()
	go func() {
		cliS.Accept(cliL)
		wg.Done()
	}()
	wg.Wait()
	return nil
}