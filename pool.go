package rpcx

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"math/rand"
	"sync"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	"github.com/meshplus/bitxhub-model/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
)

type grpcClient struct {
	broker   pb.ChainBrokerClient
	conn     *grpc.ClientConn
	nodeInfo *NodeInfo
}

type ConnectionPool struct {
	timeoutLimit time.Duration // timeout limit config for dialing grpc
	currentConn  *grpcClient
	resources    chan *grpcClient
	connections  []*grpcClient
	logger       Logger
	closed       bool
	//availableCli chan bool
	lock sync.Mutex
}

// init a connection
func NewPool(config *config) (*ConnectionPool, error) {
	pool := &ConnectionPool{
		logger:       config.logger,
		timeoutLimit: config.timeoutLimit,
		resources:    make(chan *grpcClient, config.resourcesSize),
		//availableCli: make(chan bool),
	}
	for _, nodeInfo := range config.nodesInfo {
		cli := &grpcClient{
			nodeInfo: nodeInfo,
		}
		pool.connections = append(pool.connections, cli)
	}
	return pool, nil
}

func (pool *ConnectionPool) Close() error {
	for _, c := range pool.connections {
		if c.conn == nil {
			continue
		}
		pool.closed = true
		if err := c.conn.Close(); err != nil {
			pool.logger.Errorf("stop connection with %v error: %v", c.nodeInfo.Addr, err)
			continue
		}
	}
	return nil
}

// get grpcClient will try to get idle grpcClient
func (pool *ConnectionPool) getClient() (*grpcClient, error) {
	if pool.currentConn != nil {
		for {
			select {
			case cli, ok := <-pool.resources:
				if !ok {
					return nil, fmt.Errorf("cli already closed")
				}
				//pool.logger.Infof("ge client successfully")
				return cli, nil
			default:
				pool.logger.Debugf("client available is %t,waiting cli idle\n", pool.currentConn.available())
			}
		}
	} else {
		pool.lock.Lock()
		defer pool.lock.Unlock()
		randGenerator := rand.New(rand.NewSource(time.Now().Unix()))
		if err := retry.Retry(func(attempt uint) error {
			randomIndex := randGenerator.Intn(len(pool.connections))
			cli := pool.connections[randomIndex]
			if cli.conn == nil || cli.conn.GetState() == connectivity.Shutdown {
				// try to build a connect or reconnect
				opts := []grpc.DialOption{grpc.WithBlock(), grpc.WithTimeout(pool.timeoutLimit)}
				// if EnableTLS is set, then setup connection with ca cert
				if cli.nodeInfo.EnableTLS {
					certPathByte, err := ioutil.ReadFile(cli.nodeInfo.CertPath)
					if err != nil {
						return err
					}
					cp := x509.NewCertPool()
					if !cp.AppendCertsFromPEM(certPathByte) {
						return fmt.Errorf("credentials: failed to append certificates")
					}
					cert, err := tls.LoadX509KeyPair(cli.nodeInfo.AccessCert, cli.nodeInfo.AccessKey)
					creds := credentials.NewTLS(&tls.Config{
						Certificates: []tls.Certificate{cert}, ServerName: cli.nodeInfo.CommonName, RootCAs: cp})

					if err != nil {
						pool.logger.Debugf("Creat tls credentials from %s for client %s", cli.nodeInfo.CertPath, cli.nodeInfo.Addr)
						return fmt.Errorf("%w: tls config is not right", ErrBrokenNetwork)
					}
					opts = append(opts, grpc.WithTransportCredentials(creds))
				} else {
					opts = append(opts, grpc.WithInsecure())
				}
				conn, err := grpc.Dial(cli.nodeInfo.Addr, opts...)
				if err != nil {
					pool.logger.Debugf("Dial with addr: %s fail", cli.nodeInfo.Addr)
					return fmt.Errorf("%w: dial node %s failed", ErrBrokenNetwork, cli.nodeInfo.Addr)
				}
				cli.conn = conn
				cli.broker = pb.NewChainBrokerClient(conn)
				pool.currentConn = cli
				pool.logger.Infof("Establish connection with bitxhub %s successfully", cli.nodeInfo.Addr)
				return nil
			}

			if cli.available() {
				pool.currentConn = cli
				return nil
			}
			pool.logger.Debugf("Client for %s is not usable", pool.connections[randomIndex].nodeInfo.Addr)
			return fmt.Errorf("%w: all nodes are not available", ErrBrokenNetwork)
		}, strategy.Wait(500*time.Millisecond), strategy.Limit(uint(5*len(pool.connections)))); err != nil {
			return nil, err
		}
		pool.logger.Infof("create client")
		return pool.currentConn, nil
	}
}

func (grpcCli *grpcClient) available() bool {
	s := grpcCli.conn.GetState()
	return s == connectivity.Idle || s == connectivity.Ready
}

func (cli *ChainClient) release(currentClient *grpcClient) {
	cli.pool.lock.Lock()
	defer cli.pool.lock.Unlock()
	var err error

	if cli.pool.closed {
		err = cli.pool.Close()
		cli.logger.Errorf("close pool err :%s", err)
		return
	}

	select {
	case cli.pool.resources <- currentClient:
		//cli.logger.Infof("release cli to resources")
	default:
		err = cli.pool.Close()
		cli.logger.Errorf("close pool err :%s", err)
	}
}
