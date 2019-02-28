package main

import (
	"fmt"
	"apigw-contrib/util/conf"
	"github.com/Miss-you/go-presscall/pb/tgwadm"
	"github.com/Miss-you/go-presscall/sendpb"
	"os"
	"encoding/json"
)

type OneOperationInfo struct {
	OpType	string
}

type PresscallInfo struct {
	OpList []OneOperationInfo
}

var PressConf PresscallInfo
var svrRsp tgwadm.CtrlMsg

func main() {
	//init conf
	err := conf.InitConfig("presscall.conf", &PressConf)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var req tgwadm.CtrlMsg 
	var rsp tgwadm.CtrlMsg
	ReadJson("pb.json", &req)
	fmt.Println(req)

	ReadJson("svrpb.json", &svrRsp)
	fmt.Println(svrRsp)

	//tmp, _ := json.Marshal(req)
	//fmt.Println(string(tmp))

	//ctrlMsg
	err = sendpb.SendPbReq(&req, &rsp)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("rsp: ", rsp)
	
	fmt.Println("Hello world.")
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


//tcp server.
func tcpServer() {
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
	buf := make([]byte, 2048) 
	defer conn.Close()
	for {
		_, err := sendpb.recvMsg(conn, buf)
		if err != nil {
			fmt.Println("recvConnMsg err: ", err)
			return
		}

		data, err := proto.Marshal(svrRsp)
		if err != nil {
			fmt.Println(err)
			return err
		}

		fullData := sendpb.addMagicBodySize(data)
		err = sendpb.sendMsg(conn, fullData)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}   
}