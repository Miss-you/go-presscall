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

	tmp, _ := json.Marshal(req)
	fmt.Println(string(tmp))

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
