//Package master, for send and recv data from apigw master

package sendpb


import (
    "fmt"
	"encoding/binary"
	"log"
    "github.com/golang/protobuf/proto" 
	"apigw-contrib/util/conf"
)

const MAGIC_NUM  = 0x20150812

type SendConfInfo struct {
	Ip  string `json:"ip"`
    Port uint32 `json:"port"`
}

var SendConf SendConfInfo

func init() {
	err := conf.InitConfig("sendpb.conf", &SendConf)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	fmt.Println("SendConf: ", SendConf)
}

type MasterParam struct {
    Ip  string `json:"ip"`
    Port uint32 `json:"port"`
    Data    []byte  `json:"data,omitempty"`
}
/*
type MasterSimpleResponse struct {
    Errno   int `json:"errno"`
    Errmsg  string  `json:"errmsg"`
}
*/

func AddMagicBodySize(msg []byte) []byte {
    new_msg_data := make([]byte, 8)
    msg_size := len(msg)

    fmt.Println("msg_szie", msg_size)
    
    /*htonl*/
    binary.BigEndian.PutUint32(new_msg_data[0:4], MAGIC_NUM)
    binary.BigEndian.PutUint32(new_msg_data[4:8], uint32(msg_size))

    new_msg_data = append(new_msg_data, msg...)
   
    message := fmt.Sprintf("new_msg_size:%d", len(new_msg_data)) 
    fmt.Println("message: ", message)
    fmt.Println("new_msg_data: ", new_msg_data)
    return new_msg_data
} 

func removeMagicBodySize(msg []byte) []byte {
    full_size := len(msg)
    message := fmt.Sprintf("full_size:%d", full_size)
    fmt.Println(message)

    new_msg := make([]byte, full_size - 8)
    new_msg = msg[8:]
    return new_msg
}


//func sendPbReq()
func SendPbReq(req proto.Message, rsp proto.Message) (err error) {
	data, err := proto.Marshal(req)
	if err != nil {
        fmt.Println(err)
        return err
	}

	fullData := AddMagicBodySize(data)

	conn, err := masterConnect(SendConf.Ip, SendConf.Port, 15)
    if err != nil {
        return err
    }
	defer conn.Close()
	
	err = SendMsg(conn, fullData)
    if err != nil {
        return err
	}
	
	recv_buf := make([]byte, 1024)
    recv_len, err := RecvMsg(conn, recv_buf)
    if err != nil {
        return err
	}
	
	fmt.Println("recv_len:", recv_len)
    real_msg := recv_buf[8:recv_len]
    err = proto.Unmarshal(real_msg, rsp)
    
	return nil
}
