package main

import (
	"fmt"
	//"github.com/Miss-you/go-presscall/pb/tgwadm"
	//"github.com/Miss-you/go-presscall/sendpb"
	"go-presscall/pb/tgwadm"
	"go-presscall/sendpb"
	"os"
	"encoding/json"
	"net"
	"log"
	"github.com/golang/protobuf/proto" 
)

type PresscallInfo struct {
	ClientType string
}

type ServerRspInfo struct {
	RspList []ServerPbRspInfo `json:"rsp_list"`
}

type ServerPbRspInfo struct {
	PbType uint32 `json:"type"`
	PbMsg tgwadm.CtrlMsg `json:"msg"`
}

var presscallConf PresscallInfo

func main() {
	ReadJson("presscall.json", &presscallConf)
	if presscallConf.ClientType == "client" {
		tcpClient()
	} else if presscallConf.ClientType == "server" {
		tcpServer()
	} else {
		fmt.Println("invalid ClientType. need client or server.")
	}

	return
}

func tcpClient() {
	var req tgwadm.CtrlMsg 
	var rsp tgwadm.CtrlMsg 
	ReadJson("pb.json", &req)
	fmt.Println(req)
	err := sendpb.SendPbReq(&req, &rsp)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("rsp: ", rsp)

	return
}

var rspMap map[uint32]tgwadm.CtrlMsg
//tcp server.
func tcpServer() {
	var rspList ServerRspInfo
	ReadJson("svrpb.json", &rspList)
	fmt.Println(rspList)

	rspMap = make(map[uint32]tgwadm.CtrlMsg)
	for i, _ := range rspList.RspList {
		rspMap[rspList.RspList[i].PbType] = rspList.RspList[i].PbMsg
	}

	listen_sock, err := net.Listen("tcp", ":8001")
    if err != nil {
		log.Fatal("init tcp service failed.")
	}
    defer listen_sock.Close()

    for {
        new_conn, err := listen_sock.Accept()
        if err != nil {
            continue    
        }
        go recvConnMsg(new_conn)
    }
}

func recvConnMsg(conn net.Conn) {
	var req tgwadm.CtrlMsg
	buf := make([]byte, 2048) 
	defer conn.Close()
	for {
		recvLen, err := sendpb.RecvMsg(conn, buf)
		if err != nil {
			fmt.Println("recvConnMsg err: ", err)
			return
		}

		//fmt.Println(buf)

		err = proto.Unmarshal(buf[8:recvLen], &req)
		if err != nil {
			fmt.Println("marshal req err: ", err)
			return
		}

		fmt.Println("req: ", req)

		rsp, ok := rspMap[*req.Header.CmdType]
		if !ok {
			data, err := proto.Marshal(&req)
			if err != nil {
				fmt.Println(err)
				return
			}

			fullData := sendpb.AddMagicBodySize(data)
			err = sendpb.SendMsg(conn, fullData)
			if err != nil {
				fmt.Println(err)
				return
			}
		} else {
			fmt.Println("rsp: ", rsp)
			data, err := proto.Marshal(&rsp)
			if err != nil {
				fmt.Println(err)
				return
			}

			fullData := sendpb.AddMagicBodySize(data)
			err = sendpb.SendMsg(conn, fullData)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}   
}

func ReadJson(confPath string, confVar interface{}) error {
	file, err := os.Open(confPath)
	if err != nil {
		fmt.Println("open config file failed. config file :", confPath)
		fmt.Println(err)
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&confVar)
	if err != nil {
		fmt.Println("[ReadJson]decode config failed.")
		fmt.Println(err)
		return err
	}

	{
		//print config
		tmpInter := map[string]interface{}{}
		tmpJson, _ := json.Marshal(confVar)
		json.Unmarshal(tmpJson, &tmpInter)

		//loop & print
		for field, val := range tmpInter {
			fmt.Println("KV Pair: ", field, val)
		}
	}

	return nil
}