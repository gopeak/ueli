package worker

import (
	//"math/rand"
	"morego/global"
	"morego/golog"
	"strconv"
	"strings"
	//"sync/atomic"
	"morego/protocol"
	//sync"
	"bufio"
	"fmt"
	"net"
	//"os"
	"encoding/json"
	"morego/lib/antonholmquist/jason"
	"time"
)

// 初始化worker服务
func InitWorkerServer() {

	for _, data := range global.Config.WorkerServer.Servers {

		host, _ := data[0].(string)
		port_str, _ := data[1].(string)
		worker_language, _ := data[2].(string)
		port, _ := strconv.Atoi(port_str)
		global.WorkerServers = append(global.WorkerServers, []string{host, port_str})
		fmt.Println("worker_language:", worker_language)
		if worker_language == "go" {
			go WorkerServer(host, port)
		}

	}
	//fmt.Println("global.WorkerServers:", global.WorkerServers)
}

/**
 * 监听客户端连接
 */
func WorkerServer(host string, port int) {

	fmt.Println("WorkerServer :", host, port)
	listen, err := net.ListenTCP("tcp", &net.TCPAddr{net.ParseIP(host), (port), ""})
	if err != nil {
		golog.Error("ListenTCP Exception:", err.Error())
		return
	}

	// 处理客户端连接
	for {
		conn, err := listen.AcceptTCP()
		if err != nil {
			golog.Error("AcceptTCP Exception::", err.Error(), time.Now().UnixNano())
			break
		}
		// 校验ip地址
		conn.SetKeepAlive(true)
		defer conn.Close()
		//conn.SetNoDelay(false)
		golog.Info("RemoteAddr:", conn.RemoteAddr().String())

		if global.PackSplitType == "bufferio" {
			go handleWorkerStrSplit(conn)
		}
		if global.PackSplitType == "json" {
			go handleWorkerJson(conn)
		}

	} //end for {
}

func handleWorkerStrSplit(conn *net.TCPConn) {

	//声明一个管道用于接收解包的数据
	reader := bufio.NewReader(conn)

	for {
		str, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("HandleConn connection error: ", err.Error())
			conn.Write([]byte(WrapRespErrStr(err.Error())))
			continue
		}
		//fmt.Println( "HandleWorkerStr str: ",str)
		go func(str string, conn *net.TCPConn) {

			msg_arr := strings.Split(str, "||")
			if len(msg_arr) < 5 {
				//conn.Write([]byte( WrapRespErrStr("request data length error-->"+str)))
				return
			}
			cmd := "user.getSession" //msg_arr[1];
			req_sid := msg_arr[2]
			req_id, _ := strconv.Atoi(msg_arr[3])
			data := msg_arr[4]
			resp_str := WrapRespStr(cmd, req_sid, req_id, data)
			conn.Write([]byte(resp_str))

		}(str, conn)
	}
}

func handleWorkerFlatBuffer(conn *net.TCPConn) {

	//声明一个管道用于接收解包的数据
	reader := bufio.NewReader(conn)

	for {

		str, err := reader.ReadString('\n')
		//fmt.Println("ReadString: ", str)
		if err != nil {
			//fmt.Println( "HandleConn connection error: ", err.Error())
			break
		}
		buf := []byte(str)
		go func(buf []byte, conn *net.TCPConn) {

			msg := protocol.GetRootAsData(buf, 0)
			//  do some thing
			cmd := string(msg.Cmd())
			data := string(msg.Data())
			req_sid := string(msg.Sid())
			req_id := int(msg.ReqId())
			golog.Info("handleWorker  ", cmd, data, req_sid, req_id)
			//fmt.Println("cmd: ", cmd)
			conn.Write(append(buf, '\n'))
		}(buf, conn)

	}

}

func handleWorkerJson(conn *net.TCPConn) {

	//声明一个管道用于接收解包的数据

	d := json.NewDecoder(conn)
	for {

		var msg interface{}

		err := d.Decode(&msg)
		if err != nil {

			conn.Close()
			fmt.Println("d.Decode(&msg) ", err.Error())
			break
		}
		buf, err_encode := json.Marshal(msg)
		if err_encode != nil {
			fmt.Println("json.Marshal error:", err_encode.Error())
			conn.Close()
			break
		}
		msg_json, errjson := jason.NewObjectFromBytes(buf)
		if errjson != nil {
			continue
		}
		cmd, _ := msg_json.GetString("cmd")
		token, _ := msg_json.GetString("token")
		golog.Info("handleWorkerJson:", cmd, token)

		go func(buf []byte, conn *net.TCPConn) {
			conn.Write(append(buf, '\n'))
		}(buf, conn)

	}

}

/**
 * 封包返回错误的消息
 */
func WrapRespErrStr(err string) string {
	str := fmt.Sprintf("%d||%s||%s||%d||%s", protocol.TypeError, "", "", 0, err)
	return str
}

/**
 * 封包返回数据
 */
func WrapRespStr(cmd string, from_sid string, req_id int, data string) string {
	str := fmt.Sprintf("%d||%s||%s||%d||%s", protocol.TypeReply, cmd, from_sid, req_id, data)
	return str
}
