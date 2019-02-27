package sendpb

import (
    "net"
    "fmt"
    "time"
    "errors"
    "io"
    "encoding/binary"
)

/* check package is full or not*/
func CompletePackage(msg []byte, data_len int) int {
    //fmt.Println("data_len", data_len)
    if data_len <= 8 {
        return 0
    }
    /*
    magic_byte:= msg[0:4]
    magic_num := binary.BigEndian.Uint32(magic_byte)
    if magic_num != MAGIC_NUM {
        return -1
    } 
    */

    len_byte := msg[4:8]
    parse_len := binary.BigEndian.Uint32(len_byte)
    
    if int(parse_len) <= data_len {
        return int(parse_len)
    }
    return -1
}



/*timeout is second*/
func masterConnect(host string, port uint32, timeout int)(net.Conn, error){
    var masterStr string
    masterStr = fmt.Sprintf("%s:%d", host, port)
    conn, err := net.DialTimeout("tcp", masterStr, time.Duration(timeout) * time.Second) 
    return conn, err
}


func sendMsg(p net.Conn, msg []byte) error {
    if len(msg) == 0 {
        return errors.New("send msg to master is null")
    }
    
    _, err := p.Write(msg)

    if err != nil {
        return err
    }
    return nil
}

func recvMsg(p net.Conn, msg []byte) (int, error) {
    temp := make([]byte, 1024)
    msg_len := 0
    check_result := 0
    for {
        p.SetReadDeadline(time.Now().Add(time.Second * 10))
        recv_cnt, err := p.Read(temp)
        if err != nil {
            if err != io.EOF {
                return 0, err    
            }   
            break
        }   
        copy(msg[msg_len:], temp)  

        msg_len = msg_len + recv_cnt
        check_result = CompletePackage(msg, msg_len)
        if check_result > 0 { 
            break;
        }   
    }   
                                                                                                                                                                
    if check_result <= 0 { 
        return 0, errors.New("recv package uncomplete")    
    }   
    return msg_len, nil 

}
