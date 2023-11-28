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
	"github.com/bobwong89757/cellnet/rpc"
	"github.com/bobwong89757/cellnet/util"
	"github.com/bobwong89757/golog/logs"
)

const peerAddress = "10.0.40.20:12345"

func main() {
	customLogger := logs.NewLogger()
	customLogger.EnableFuncCallDepth(true)
	log.SetLog(customLogger)
	queue := cellnet.NewEventQueue()

	//timer.NewLoop(nil,time.Millisecond * 100,func(loop *timer.Loop) {
	//	fmt.Println(fmt.Sprintf("当前协程总数:%d",runtime.NumGoroutine()) )
	//},nil).Start()

	peerIns := peer.NewGenericPeer("kcp.Acceptor", "server", peerAddress, queue)

	proc.BindProcessorHandler(peerIns, "kcp.ltv", func(ev cellnet.Event) {

		switch msg := ev.Message().(type) {
		case *cellnet.SessionAccepted: // 接受一个连接
			addr, has := util.GetRemoteAddrss(ev.Session())
			log.GetLog().Debugf("server accepted %s,%v", addr, has)
		case *proto.EchoAck: // 收到连接发送的消息

			fmt.Printf("server recv %+v\n", msg)

			ack := &proto.EchoAck{
				Msg: msg.Msg,
				Ext: "5",
			}

			// 当服务器收到的是一个rpc消息
			if rpcevent, ok := ev.(*rpc.RecvMsgEvent); ok {

				// 以RPC方式回应
				rpcevent.Reply(ack)
			} else {

				// 收到的是普通消息，回普通消息
				ev.Session().Send(ack)
				ev.Session().Close()
			}

		case *cellnet.SessionClosed: // 连接断开
			fmt.Println("session closed: ", ev.Session().ID())
		}

	})

	peerIns.Start()

	queue.StartLoop()

	queue.Wait()
}
