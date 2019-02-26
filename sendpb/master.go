//Package master, for send and recv data from apigw master

package master


import (
    "fmt"
    "encoding/binary"
	"errors"
	"log"
    "github.com/golang/protobuf/proto" 
    
	"apigw-oss/pb/apigw_common"
    "apigw-oss/pb/apigw_master"
    "apigw-oss/common"
	"apigw-oss/logger"
	"apigw-contrib/util/conf"
)

const MAGIC_NUM  = 0x20150812

type SendConfInfo struct {
	Ip  string `json:"ip"`
    Port uint32 `json:"port"`
}

var SendConf SendConfInfo

func init() {
	err := conf.InitConfig("sendpb.conf", SendConf)
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

func addMagicBodySize(msg []byte) []byte {
    new_msg_data := make([]byte, 8)
    msg_size := len(msg)

    logger.MasterSugar.Info("msg_szie", msg_size)
    
    /*htonl*/
    binary.BigEndian.PutUint32(new_msg_data[0:4], MAGIC_NUM)
    binary.BigEndian.PutUint32(new_msg_data[4:8], uint32(msg_size))

    new_msg_data = append(new_msg_data, msg...)
   
    message := fmt.Sprintf("new_msg_size:%d", len(new_msg_data)) 
    logger.MasterSugar.Info(message)
    return new_msg_data
} 

func removeMagicBodySize(msg []byte) []byte {
    full_size := len(msg)
    message := fmt.Sprintf("full_size:%d", full_size)
    logger.MasterSugar.Info(message)

    new_msg := make([]byte, full_size - 8)
    new_msg = msg[8:]
    return new_msg
}
// notice: the full message = htonl(magic_num) + body_length + msg

func analysisCommonResponse(msg []byte) error {
    new_msg := removeMagicBodySize(msg) 
    response := &apigw_master.ApigwMasterMsg{}
    
    message := fmt.Sprintf("analysisCommonResponse msg_size:%d", len(new_msg))
    logger.MasterSugar.Info(message)

    err := proto.Unmarshal(new_msg, response)
    
    if err != nil {
        message = fmt.Sprintf("unmarshal error: %s", err) 
        logger.MasterSugar.Error(message)
    } 
   
    common_rsp := response.GetCommonRsp()
    retmsg := common_rsp.GetRetmsg()
    errno := retmsg.GetRetcode()
    errmsg := retmsg.GetRetmsg()

    message = fmt.Sprintf("errno:%d errmsg:%s", errno, errmsg)
    logger.MasterSugar.Info(message)

    if errno != 0 {
        return errors.New(errmsg)
    }
    return nil
}



func constructSetEnvConfMessage(serviceEnvList []common.ApigwServiceEnvInfo) ([]byte, error) {

    /*use the first's setid */
    set_id := serviceEnvList[0].SetId 
    msg := &apigw_master.ApigwMasterMsg{}
    var seq uint64
    seq = (uint64(set_id) << 32) | uint64(common.GetRandInt())
    header := &apigw_common.Head {
        VersionHigh :  proto.Uint32(0),
        VersionLow :   proto.Uint32(0),
        Seq :           proto.Uint64(seq),
        CmdType:       proto.Uint32(uint32(apigw_master.MasterCmd_SET_ENV_CONF_CMD)),
    }

    var req apigw_master.SetEnvConfigReq

    for _, serviceEnv := range serviceEnvList {
        var pbServiceEnv *apigw_common.ApiServiceRunningEnv
        
        pbServiceEnv = common.ApigwServiceEnvInfoToPb(&serviceEnv)
        req.EnvList = append(req.EnvList, pbServiceEnv)
    }

    msg.Header = header
    msg.SetEnvConfigReq = &req

    data, err := proto.Marshal(msg)
    if err != nil {
        return nil, err
    }

    full_data := addMagicBodySize(data)   
    return full_data, err 
}

//MasterCmd_SET_ENV_CONF_CMD
func SendSetEnvConfRequest(masterParam MasterParam, serviceEnvList []common.ApigwServiceEnvInfo) error {
    data, err := constructSetEnvConfMessage(serviceEnvList)
    if err != nil {
        logger.MasterSugar.Error(err)
        return err
    }
    logger.MasterSugar.Info("send env size:", len(serviceEnvList))
    conn, err := masterConnect(masterParam.Ip, masterParam.Port, 15)
    if err != nil {
        return err
    }
    defer conn.Close()
    
    err = sendMsg(conn, data)
    if err != nil {
        return err
    }

    recv_buf := make([]byte, 1024)
    recv_len, err := recvMsg(conn, recv_buf)
    if err != nil {
        return err
    }
    
    logger.MasterSugar.Info("recv_len:", recv_len)
    real_msg := recv_buf[:recv_len]
    err = analysisCommonResponse(real_msg)
    if err != nil {
        return err
    }

    return nil
}
