package main

import (
	"fmt"
	"github.com/ligato/networkservicemesh/controlplane/pkg/apis/crossconnect"
	lcon "github.com/ligato/networkservicemesh/controlplane/pkg/apis/local/connection"
	rcon "github.com/ligato/networkservicemesh/controlplane/pkg/apis/remote/connection"
	"github.com/ligato/networkservicemesh/controlplane/pkg/model"
	"github.com/ligato/networkservicemesh/controlplane/pkg/monitor_crossconnect_server"
	"github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Capture signals to cleanup before exiting
	c := make(chan os.Signal, 1)
	signal.Notify(c,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	myModel := model.NewModel("127.0.0.1:5000")
	crossConnectAddress := "127.0.0.1:5007"

	err, grpcServer, monitor := monitor_crossconnect_server.StartNSMCrossConnectServer(myModel, crossConnectAddress)

	if err != nil {
		logrus.Fatalf("Error starting crossconnect server: %s", err)
	}

	go feedFakeCrossConnect(monitor, 5000, 10, 0.5)

	select {
	case <-c:
		logrus.Info("Closing dataplane registration")
		grpcServer.Stop()
	}
}

// A struct to hold fake xcons
// either local == nil or both remoteSender, remoteReceiver == nil
type crossConnectHolder struct {
	local          *crossconnect.CrossConnect
	remoteSender   *crossconnect.CrossConnect
	remoteReceiver *crossconnect.CrossConnect
}

func genLocalConnection(pId *int) *crossConnectHolder {
	defer func() {
		*pId++
	}()
	return &crossConnectHolder{
		local: &crossconnect.CrossConnect{
			Id:      fmt.Sprintf("%x", *pId),
			Payload: "Ethernet",
			Source: &crossconnect.CrossConnect_LocalSource{
				&lcon.Connection{
					Id:             fmt.Sprintf("%x0", *pId),
					NetworkService: fmt.Sprintf("%x", rand.Intn(1048576)),
					Mechanism: &lcon.Mechanism{
						Type: lcon.MechanismType(rand.Int31n(6)),
						Parameters: map[string]string{
							"para_key1": "para_val1",
							"para_key2": "para_val2",
						},
					},
					Context: map[string]string{
						"ctx_key1": "ctx_val1",
						"ctx_key2": "ctx_val2",
					},
					Labels: map[string]string{
						"lbl_key1": "lbl_val1",
						"lbl_key2": "lbl_val2",
					},
				},
			},
			Destination: &crossconnect.CrossConnect_LocalDestination{
				&lcon.Connection{
					Id:             fmt.Sprintf("%x1", *pId),
					NetworkService: fmt.Sprintf("%x", rand.Intn(1048576)),
					Mechanism: &lcon.Mechanism{
						Type: lcon.MechanismType(rand.Int31n(6)),
						Parameters: map[string]string{
							"para_key1": "para_val1",
							"para_key2": "para_val2",
						},
					},
					Context: map[string]string{
						"ctx_key1": "ctx_val1",
						"ctx_key2": "ctx_val2",
					},
					Labels: map[string]string{
						"lbl_key1": "lbl_val1",
						"lbl_key2": "lbl_val2",
					},
				},
			},
		},
		remoteSender:   nil,
		remoteReceiver: nil,
	}
}

func genRemoteConnection(pId *int) *crossConnectHolder {
	defer func() {
		*pId += 2
	}()

	remoteConnection := &rcon.Connection{
		Id:             fmt.Sprintf("%x1", *pId), // here use the sender ID
		NetworkService: fmt.Sprintf("%x", rand.Intn(1048576)),
		Mechanism: &rcon.Mechanism{
			Type: rcon.MechanismType(rand.Int31n(6)),
			Parameters: map[string]string{
				"para_key1": "para_val1",
				"para_key2": "para_val2",
			},
		},
		Context: map[string]string{
			"ctx_key1": "ctx_val1",
			"ctx_key2": "ctx_val2",
		},
		Labels: map[string]string{
			"lbl_key1": "lbl_val1",
			"lbl_key2": "lbl_val2",
		},
		SourceNetworkServiceManagerName:      "NSMD_0",
		DestinationNetworkServiceManagerName: "NSMD_1",
	}

	return &crossConnectHolder{
		local: nil,
		remoteSender: &crossconnect.CrossConnect{
			Id:      fmt.Sprintf("%x", *pId),
			Payload: "Ethernet",
			Source: &crossconnect.CrossConnect_LocalSource{
				&lcon.Connection{
					Id:             fmt.Sprintf("%x0", *pId),
					NetworkService: fmt.Sprintf("%x", rand.Intn(1048576)),
					Mechanism: &lcon.Mechanism{
						Type: lcon.MechanismType(rand.Int31n(6)),
						Parameters: map[string]string{
							"para_key1": "para_val1",
							"para_key2": "para_val2",
						},
					},
					Context: map[string]string{
						"ctx_key1": "ctx_val1",
						"ctx_key2": "ctx_val2",
					},
					Labels: map[string]string{
						"lbl_key1": "lbl_val1",
						"lbl_key2": "lbl_val2",
					},
				},
			},
			Destination: &crossconnect.CrossConnect_RemoteDestination{
				remoteConnection,
			},
		},
		remoteReceiver: &crossconnect.CrossConnect{
			Id:      fmt.Sprintf("%x", *pId+1),
			Payload: "Ethernet",
			Source: &crossconnect.CrossConnect_RemoteSource{
				remoteConnection,
			},
			Destination: &crossconnect.CrossConnect_LocalDestination{
				&lcon.Connection{
					Id:             fmt.Sprintf("%x2", *pId+1),
					NetworkService: fmt.Sprintf("%x", rand.Intn(1048576)),
					Mechanism: &lcon.Mechanism{
						Type: lcon.MechanismType(rand.Int31n(6)),
						Parameters: map[string]string{
							"para_key1": "para_val1",
							"para_key2": "para_val2",
						},
					},
					Context: map[string]string{
						"ctx_key1": "ctx_val1",
						"ctx_key2": "ctx_val2",
					},
					Labels: map[string]string{
						"lbl_key1": "lbl_val1",
						"lbl_key2": "lbl_val2",
					},
				},
			},
		},
	}
}

const (
	update = iota
	delete
)

const (
	remote = iota
	local
)

const seed = 135797531 // 16^5

// maxCrossConnectCount: int
// 	 max count of fake connections
//
// remoteConnectionPossibility: float32
//   possibility of connection type,
//   0 for all local connections, 1 for all remote connections
func feedFakeCrossConnect(
	monitor monitor_crossconnect_server.MonitorCrossConnectServer,
	milliSecondInterval int, maxCrossConnectCount int, remoteConnectionPossibility float32) {

	rand.Seed(seed)
	// A buffer slice working as a vector
	// Tricks to make slice a vector: https://github.com/golang/go/wiki/SliceTricks
	crossConnectBuffer := make([]*crossConnectHolder, 0, maxCrossConnectCount)
	nLocal := 0
	nRemote := 0

	id := 0

	var behavior int
	// work until process be killed
	for {
		switch nTotal := nLocal + nRemote; nTotal {
		case 0:
			behavior = update
		case maxCrossConnectCount:
			behavior = delete
		default:
			behavior = rand.Intn(2)
		}

		if behavior == update {
			var con *crossConnectHolder
			switch cmp := rand.Float32() > remoteConnectionPossibility; cmp {
			case true: // local
				nLocal++
				con = genLocalConnection(&id)
				crossConnectBuffer = append(crossConnectBuffer, con)
				monitor.UpdateCrossConnect(con.local)
				logrus.Info("Generate a Local CrossConnect")
			case false: // remote
				nRemote++
				con = genRemoteConnection(&id)
				crossConnectBuffer = append(crossConnectBuffer, con)
				monitor.UpdateCrossConnect(con.remoteReceiver)
				monitor.UpdateCrossConnect(con.remoteSender)
				logrus.Info("Generate a pair of Remote CrossConnects")
			}
		} else {
			i := rand.Intn(len(crossConnectBuffer))
			con := crossConnectBuffer[i]
			switch con.local == nil {
			case true: // local
				nLocal--
				monitor.DeleteCrossConnect(con.local)
				logrus.Info("Delete a Local CrossConnect")
			case false: // remote
				nRemote--
				monitor.DeleteCrossConnect(con.remoteSender)
				monitor.DeleteCrossConnect(con.remoteReceiver)
				logrus.Info("Delete a pair of Remote CrossConnects")
			}
			// remove the xconHolder from buffer
			copy(crossConnectBuffer[i:], crossConnectBuffer[i+1:])
			crossConnectBuffer[len(crossConnectBuffer)-1] = nil
			crossConnectBuffer = crossConnectBuffer[:len(crossConnectBuffer)-1]
		}
		logrus.WithFields(logrus.Fields{"Local Xcon": nLocal, "Remote Xcon": nRemote}).Info("Buffer Overview")
		logrus.WithFields(logrus.Fields{"Interval (seconds)": milliSecondInterval / 1000}).Info("Zzz... Zzzzz.....")
		time.Sleep(time.Duration(milliSecondInterval) * time.Millisecond)
	}
}
