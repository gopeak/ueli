package golang

import (
	"morego/golog"
	"github.com/robfig/cron"
	"morego/global"
	z_type "morego/type"
	"fmt"
	"time"
	"net"
	"morego/protocol"
	"morego/hub"
	"bufio"
	"encoding/json"
	"morego/lib/syncmap"
	"morego/lib/conn_pool/pool"
	"morego/util"
	"strconv"
)


type Sdk struct {


	Connected bool

	HubConn *net.TCPConn

	Cmd string

	Sid string

	Reqid int

	Data interface{}

}

type PushReqHub struct {

	Sid bool
	Msg string
	Info map[string]string

}

type AfterWorkCallback func(   resp_buf string ) (string)

var ReqSeqCaalbacks *syncmap.SyncMap

var HubConnsPool  pool.Pool

var ReqHubConns  =  make( []*net.TCPConn, 0 )
var InitialCap  int

func (this *Sdk) Init(cmd string,sid string,reqid int,data interface{}) *Sdk{

	this.Cmd = cmd
	this.Sid = sid
	this.Reqid = reqid
	this.Data = data
	this.Connected = false
	return this
}

// 数据连接
func (this *Sdk) connect() bool{

	if this.HubConn!=nil {
		return true
	}
	data :=  global.Config.WorkerServer.ToHub
	hub_host := data[0]
	hub_port_str := data[1]
	ip_port := hub_host + ":" + hub_port_str

	tcpAddr, _ := net.ResolveTCPAddr("tcp4", ip_port)
	hubconn, err_req := net.DialTCP("tcp", nil, tcpAddr)
	if( err_req!=nil ){
		this.HubConn=nil
		return false
	}
	this.HubConn = hubconn
	return true

}



func   InitReqHubPool() {

	// create a factory() to be used with channel based pool
	ReqSeqCaalbacks = syncmap.New()

	InitialCap  = 10

	factory    := func() (*net.TCPConn, error) {
		data :=  global.Config.WorkerServer.ToHub
		hub_host := data[0]
		hub_port_str := data[1]
		ip_port := hub_host + ":" + hub_port_str

		tcpAddr, _ := net.ResolveTCPAddr("tcp4", ip_port)
		hubconn, err_req := net.DialTCP("tcp", nil, tcpAddr)
		//fmt.Println( "InitConnectionHubPool hubconn ", hubconn )

		return hubconn,err_req
	}
	for i := 0; i < InitialCap; i++ {

		var err_req error
		conn, err_req:= factory()
		if( err_req!=nil ) {
			golog.Error( "InitConnectionHubPool hubconn  err:", err_req.Error() )
			continue
		}
		ReqHubConns = append( ReqHubConns, conn )
		go handleReqHubResponse( conn )
	}
}


// 侦听Hub server返回的数据，然后回调worker的函数
func  handleReqHubResponse(conn *net.TCPConn) {

	time.Sleep( 2*time.Second)
	reader := bufio.NewReader(conn)
	defer func() {
		err := recover()
		if err != nil {
			conn.Close()
			fmt.Println( "ReadHubResp err :", err)
		}
	}()
	for {
		buf ,err := protocol.Unpack( reader)
		if err != nil {
			fmt.Println( "Hub handleWorker connection error: ", err.Error())
			conn.Close()
			break
		}
		resp_cmd,req_id,resp_err,msg_data := protocol.ReadHubResp(buf)
		if resp_err!=""{
			golog.Error( "handleReqHubResponse ReadHubResp err:",resp_err )
			continue
		}
		callback_key:=resp_cmd + req_id
		fmt.Println( "callback_key:", callback_key )

		_item,ok := ReqSeqCaalbacks.Get( callback_key )
		if( ok ) {
			callback := _item.( AfterWorkCallback )
			fmt.Println( "callback func :", callback  )
			callback( string(msg_data) )
			ReqSeqCaalbacks.Delete( callback_key )
		}
	}

}


// 向Hub请求数据并监听返回,该请求将会阻塞除非等待返回超时
func (this *Sdk) ReqHubAsync( req_cmd string , data string ,handler AfterWorkCallback  ) (string,bool) {

	req_id := strconv.FormatInt( time.Now().UTC().UnixNano(), 10)
	req_buf := protocol.MakeHubReq( req_cmd, this.Sid, req_id, data )
	req_buf,_ = protocol.Packet( req_buf )

	index := util.RandInt64(0, int64(len(ReqHubConns)))
	req_hub_conn  := ReqHubConns[index]

	//req_hub_conn,err := HubConnsPool.Get()
	fmt.Println( "ReqHubConns:", ReqHubConns )
	if( req_hub_conn==nil  ){
		golog.Error( "req_hub_conn is nil "  )
		return "", false
	}
	callback_key:=req_cmd+ req_id
	ReqSeqCaalbacks.Set( callback_key, handler )
	fmt.Println( "ReqHubAsync:", callback_key )
	_,err := req_hub_conn.Write( req_buf )
	if err!=nil {
		golog.Error( "req_hub_conn.Write err:" , err.Error() )
	}
	return "",false
}


// 向Hub请求数据并监听返回,该请求将会阻塞除非等待返回超时
func (this *Sdk) ReqHub( req_cmd string , data string ) (string,bool) {

	req_buf := protocol.MakeHubReq( req_cmd, this.Sid, strconv.Itoa( int(this.Reqid) ), data )
	req_buf,_ = protocol.Packet( req_buf )
	fmt.Println( "req_str:", string(req_buf) )

	index := util.RandInt64(0, int64(len(ReqHubConns)))
	req_hub_conn  := ReqHubConns[index]
	req_hub_conn.Write( req_buf )
	return "",false
}

func (this *Sdk) PushHub( req_cmd string , data string ) bool {

	req_buf := protocol.MakeHubReq( req_cmd,this.Sid, strconv.Itoa( int(this.Reqid) ), data )
	req_buf,_ = protocol.Packet( req_buf )
	this.connect()
	_,err:=this.HubConn.Write( req_buf )

	if( err!=nil ) {
		return false
	}

	return true
}



// 获取服务器的根路径
func (this *Sdk)  GetBase() string {

	// 单机模式直接返回内存中数据
	if( global.SingleMode ) {
		api := new(hub.Api)
		return api.GetBase()
	}

	ret,ok :=this.ReqHub( "GetBase","" )
	if ok {
		return ret
	}
	return ""

}

// 获取服务启用状态
func (this *Sdk) GetEnableStatus() bool {

	if( global.SingleMode ) {
		api := new(hub.Api)
		return api.GetEnableStatus()
	}

	ret,ok:= this.ReqHub( "GetEnableStatus","" )
	if( !ok ){
		return false
	}
	if( ret=="1" ){
		return true
	}else{
		return false
	}

}

func (this *Sdk) Enable() bool {

	if( global.SingleMode ) {
		global.AppConfig.Enable = 1
		return true
	}
	return this.PushHub( "Enable","")


}

func (this *Sdk) Disable() bool {

	if( global.SingleMode ) {
		global.AppConfig.Enable = 0
		return true
	}
	return this.PushHub( "Disable","")

}

func (this *Sdk) AddCron(expression string, exefnc func()) bool {

	if cron, ok := global.Crons[expression]; ok {
		golog.Info("cron exist :", cron)
		return false
	}
	c := cron.New()
	c.AddFunc(expression, exefnc)
	c.Start()
	global.Crons[expression] = c
	return true

}

func (this *Sdk) RemoveCron(expression string) bool {

	if cron, ok := global.Crons[expression]; ok {
		delete(global.Crons, expression)
		cron.Stop()
	} else {
		return false
	}

	return true

}

func (this *Sdk) Get(key string) string {

	if( global.SingleMode ) {
		str,err:=hub.Get(key)
		if err!=nil {
			golog.Error("Redis Get err:",err.Error())
			return ""
		}
		return str
	}

	ret,ok := this.ReqHub( "Get",key )
	if( !ok ) {
		return ""
	}
	return ret

}

func (this *Sdk) Set(key string, value string,expire int) bool {

	if( global.SingleMode ) {
		ret,err:=hub.Set(key,value,expire)
		if err!=nil {
			golog.Error("Redis Set err:",err.Error())
			return false
		}
		return ret
	}
	json:=fmt.Sprintf(`{"key":"%s","value":"%s","expire":%d}`,key,value,expire)
	ret:= this.PushHub( "Set",json )
	return ret
}

// 该方法仅在单机模式下调用
func (this *Sdk) GetSessionType(sid string) *z_type.Session  {

	session,exist := global.SyncUserSessions.Get(sid)
	if !exist {
		return nil
	}
	return session.(*z_type.Session)
}

func (this *Sdk) GetSession(sid string)  string {

	if( global.SingleMode ) {
		api := new(hub.Api)
		return api.GetSession( sid )
	}
	ret,ok := this.ReqHub( "GetSession",sid )
	if !ok{
		return ""
	}
	return ret

}

func (this *Sdk) Kick(sid string) bool {

	if( global.SingleMode ) {
		api := new(hub.Api)
		return api.Kick( sid )
	}
	return this.PushHub( "Kick",sid)
}

func (this *Sdk) CreateChannel(id string, name string) bool {

	if( global.SingleMode ) {
		api := new(hub.Api)
		return api.CreateChannel( id,name )
	}
	json:=fmt.Sprintf(`{"id":"%s","name":"%s","expire":%d}`,id,name)
	return this.PushHub( "CreateChannel",json)

}

func (this *Sdk) RemoveChannel(id string) bool {

	if( global.SingleMode ) {
		api := new(hub.Api)
		return api.RemoveChannel( id )
	}
	return this.PushHub( "RemoveChannel",id)
}

func (this *Sdk) GetChannels() string {

	if( global.SingleMode ) {
		api := new(hub.Api)
		return api.GetChannels(  )
	}
	ret,ok := this.ReqHub( "GetChannels","" )
	if( !ok ) {
		return "{}"
	}
	return ret
}



func (this *Sdk) GetSidsByChannel(channel_id string) string {

	if( global.SingleMode ) {
		api := new(hub.Api)
		return api.GetSidsByChannel( channel_id )
	}
	ret,ok :=  this.ReqHub( "GetSidsByChannel",channel_id )
	if( !ok ) {
		return "{}"
	}
	return ret

}

func (this *Sdk) ChannelAddSid(sid string, area_id string) bool {

	if( global.SingleMode ) {
		api := new(hub.Api)
		return api.ChannelAddSid( sid, area_id  )
	}
	json:=fmt.Sprintf(`{"sid":"%s","area_id":"%s"}`,sid, area_id )
	return this.PushHub( "ChannelAddSid",json)

}

func (this *Sdk) ChannelKickSid( sid string, area_id string) bool {

	if( global.SingleMode ) {
		api := new(hub.Api)
		return api.ChannelKickSid( sid, area_id  )
	}
	json:=fmt.Sprintf(`{"sid":"%s","area_id":"%s"}`,sid, area_id )
	return this.PushHub( "ChannelKickSid",json)

}

func (this *Sdk) Push( from_sid string ,to_sid string , data  map[string]interface{} ) bool {

	data["from_sid"] = from_sid
	data["to_sid"] = to_sid
	json,_:= json.Marshal( data )
	if( global.SingleMode ) {
		api := new(hub.Api)
		return api.Push ( from_sid,to_sid, string(json)  )
	}

	return this.PushHub( "Push",string(json) )

}

func (this *Sdk) PushBySids(from_sid string,to_sids []string, data  map[string]interface{}) bool {

	for _,to_sid:=   range to_sids {
		this.Push(from_sid, to_sid, data )
	}
	return true

}

func (this *Sdk) Broatcast(sid string ,area_id string,  data  map[string]interface{} ) bool {

	data["sid"] = sid
	data["area_id"] = area_id
	json,_:= json.Marshal( data )
	if( global.SingleMode ) {
		api := new(hub.Api)
		return api.Broadcast( sid,area_id,  string(json)  )
	}
	return this.PushHub( "Broatcast",string(json) )

}


func (this *Sdk) BroadcastAll(msg string) bool {

	if( global.SingleMode ) {
		api := new(hub.Api)
		return api.BroadcastAll(   msg  )
	}
	return this.PushHub( "BroadcastAll",msg)

}


func (this *Sdk) UpdateSession( sid string, data string ) bool {

	if( global.SingleMode ) {
		api := new(hub.Api)
		return api.UpdateSession( sid, data )
	}
	json:=fmt.Sprintf(`{"sid":"%s","data":"%s"}`,sid, data )
	return this.PushHub( "UpdateSession",json)

}

func (this *Sdk)GetUserJoinedChannel(sid string) string {

	// 单机模式直接返回内存中数据
	if( global.SingleMode ) {
		api := new(hub.Api)
		return api.GetUserJoinedChannel(sid)
	}

	ret,ok :=this.ReqHub( "GetUserJoinedChannel",sid)
	if ok {
		return ret
	}
	return ""

}

func (this *Sdk)GetAllSession( ) string {

	// 单机模式直接返回内存中数据
	if( global.SingleMode ) {
		api := new(hub.Api)
		return api.GetAllSession()
	}

	ret,ok :=this.ReqHub( "GetAllSession","")
	if ok {
		return ret
	}
	return ""

}
