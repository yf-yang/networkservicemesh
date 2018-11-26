package main

import (
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ligato/networkservicemesh/controlplane/pkg/apis/crossconnect"
	lcon "github.com/ligato/networkservicemesh/controlplane/pkg/apis/local/connection"
	"github.com/ligato/networkservicemesh/controlplane/pkg/model"
	"github.com/ligato/networkservicemesh/controlplane/pkg/monitor_crossconnect_server"
	"github.com/sirupsen/logrus"
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

	if len(os.Args[1:]) == 3 {
		i, err := strconv.Atoi(os.Args[3])
		if err != nil {
			logrus.Fatalf("Error converting %s to int", os.Args[3])
		}
		go genLocalConnection(monitor, os.Args[1], os.Args[2], i)
	} else {
		logrus.Fatalf("Error: missings args")
	}
	select {
	case <-c:
		grpcServer.Stop()
	}
}

func genLocalConnection(
	monitor monitor_crossconnect_server.MonitorCrossConnectServer,
	srcInode string,
	dstInode string,
	interval int) {

	behavior := crossconnect.CrossConnectEventType_UPDATE
	for {
		cc := &crossconnect.CrossConnect{
			Id:      "cc_id",
			Payload: "Ethernet",
			Source: &crossconnect.CrossConnect_LocalSource{
				&lcon.Connection{
					Id:             "c_src_id",
					NetworkService: "ns_id",
					Mechanism: &lcon.Mechanism{
						Type: lcon.MechanismType(rand.Int31n(6)),
						Parameters: map[string]string{
							"inode": srcInode,
						},
					},
					Context: map[string]string{
						"ctx_key1": "ctx_val1",
					},
					Labels: map[string]string{
						"lbl_key1": "lbl_val1",
					},
				},
			},
			Destination: &crossconnect.CrossConnect_LocalDestination{
				&lcon.Connection{
					Id:             "c_dst_id",
					NetworkService: "ns_id",
					Mechanism: &lcon.Mechanism{
						Type: lcon.MechanismType(rand.Int31n(6)),
						Parameters: map[string]string{
							"inode": dstInode,
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
		}

		switch behavior {
		case crossconnect.CrossConnectEventType_UPDATE:
			logrus.Info("Sending update")
			monitor.UpdateCrossConnect(cc)
			behavior = crossconnect.CrossConnectEventType_DELETE
		case crossconnect.CrossConnectEventType_DELETE:
			logrus.Info("Sending delete")
			monitor.DeleteCrossConnect(cc)
			behavior = crossconnect.CrossConnectEventType_UPDATE
		}

		time.Sleep(time.Duration(interval) * time.Millisecond)
	}

}
