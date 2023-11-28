package main

import (
	"fmt"
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/examples/kcp/proto"
	"github.com/bobwong89757/cellnet/log"
	"github.com/bobwong89757/cellnet/peer"
	_ "github.com/bobwong89757/cellnet/peer/kcp"
	"github.com/bobwong89757/cellnet/proc"
	_ "github.com/bobwong89757/cellnet/proc/kcp"
	"github.com/bobwong89757/cellnet/util"
	"github.com/bobwong89757/golog/logs"
)

const peerAddress = "10.0.40.20:8902"

func main() {
	customLogger := logs.NewLogger()
	customLogger.EnableFuncCallDepth(true)
	log.SetLog(customLogger)
	queue := cellnet.NewEventQueue()
	peerIns := peer.NewGenericPeer("kcp.Connector", "server", peerAddress, queue)

	proc.BindProcessorHandler(peerIns, "kcp.ltv", func(ev cellnet.Event) {

		switch msg := ev.Message().(type) {
		case *cellnet.SessionConnected: // 接受一个连接
			addr,has := util.GetRemoteAddrss(ev.Session())
			log.GetLog().Debugf("server connect %s,%v",addr,has)

			req := &proto.LoginServer{
				UserId: "10026",
				GameToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMDAyNiwicmVtb3RlX2lwIjoiOjg4MjgiLCJleHAiOjE2MTU0NTU3MTYsImlzcyI6ImJvYiJ9.NVl7p-m94QZIRI6Kt359IexPR4PPi1BfeVMMsUU2lic",
				GameSvcID: "game#0@dev",
			}
			ev.Session().Send(req)

			//ack := &proto.EchoAck{
			//	Msg:   "goodjob",
			//	Ext: "4",
			//}
			//ev.Session().Send(ack)
		case *proto.EchoAck: // 收到连接发送的消息

			fmt.Printf("server recv %+v\n", msg)

			//ack := &tests.TestEchoACK{
			//	Msg:   msg.Msg,
			//	Value: msg.Value,
			//}
			//
			//// 当服务器收到的是一个rpc消息
			//if rpcevent, ok := ev.(*rpc.RecvMsgEvent); ok {
			//
			//	// 以RPC方式回应
			//	rpcevent.Reply(ack)
			//} else {
			//
			//	// 收到的是普通消息，回普通消息
			//	ev.Session().Send(ack)
			//}
		case *proto.LoginServerACK:
			fmt.Printf("server recv %+v\n", msg)
		case *cellnet.SessionClosed: // 连接断开
			fmt.Println("session closed: ", ev.Session().ID())
		}

	})

	peerIns.Start()

	queue.StartLoop()

	queue.Wait()
}
