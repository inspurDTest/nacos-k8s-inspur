package nacosClient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type INacosClient interface {
}

type NacosClient struct {
	logger     log.Logger
	httpClient http.Client
}

// V1
/*type ServersInfo struct {
	Servers []struct {
		IP         string `json:"ip"`
		Port       int    `json:"port"`
		State      string `json:"state"`
		ExtendInfo struct {
			LastRefreshTime int64 `json:"lastRefreshTime"`
			RaftMetaData    struct {
				MetaDataMap struct {
					NamingPersistentService struct {
						Leader          string   `json:"leader"`
						RaftGroupMember []string `json:"raftGroupMember"`
						Term            int      `json:"term"`
					} `json:"naming_persistent_service"`
				} `json:"metaDataMap"`
			} `json:"raftMetaData"`
			RaftPort string `json:"raftPort"`
			Version  string `json:"version"`
		} `json:"extendInfo"`
		Address       string `json:"address"`
		FailAccessCnt int    `json:"failAccessCnt"`
	} `json:"servers"`
}
*/
type ServersInfo struct {
	Servers []struct {
		IP         string `json:"ip"`
		Port       int    `json:"port"`
		State      string `json:"state"`
		ExtendInfo struct {
			LastRefreshTime int64 `json:"lastRefreshTime"`
			RaftMetaData    struct {
				MetaDataMap struct {
					NamingPersistentService struct {
						Leader          string   `json:"leader"`
						RaftGroupMember []string `json:"raftGroupMember"`
						Term            int      `json:"term"`
					} `json:"naming_persistent_service_v2"`
				} `json:"metaDataMap"`
			} `json:"raftMetaData"`
			RaftPort string `json:"raftPort"`
			Version  string `json:"version"`
		} `json:"extendInfo"`
		Address       string `json:"address"`
		FailAccessCnt int    `json:"failAccessCnt"`
	} `json:"data"`
}

func (c *NacosClient) GetClusterNodes(ip string) (ServersInfo, error) {
	servers := ServersInfo{}
	//增加支持ipV6 pod状态探测
	var resp *http.Response
	var err error

	// 检查每个pod控制台是否可用
	if strings.Contains(ip, ":") {
		resp, err = c.httpClient.Get(fmt.Sprintf("http://[%s]:8848/nacos/", ip))
	} else {
		resp, err = c.httpClient.Get(fmt.Sprintf("http://%s:8848/nacos/", ip))
	}
	if resp ==  nil ||  !strings.Contains(resp.Status,"200"){
		return servers, err
	}

	// 检查每个pod的状态
	if strings.Contains(ip, ":") {
		resp, err = c.httpClient.Get(fmt.Sprintf("http://[%s]:8848/nacos/v2/core/cluster/node/list?state=UP", ip))
	} else {
		resp, err = c.httpClient.Get(fmt.Sprintf("http://%s:8848/nacos/v2/core/cluster/node/list?state=UP", ip))
	}

	if err != nil {
		return servers, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return servers, err
	}
	err = json.Unmarshal(body, &servers)
	if err != nil {
		fmt.Printf("%s\n", body)
		return servers, fmt.Errorf(fmt.Sprintf("instance: %s ; %s ;body: %v", ip, err.Error(), string(body)))
	}
	return servers, nil
}

//func (c *CheckClient) getClusterNodesStaus(ip string) (bool, error) {
//	str, err := c.getClusterNodes(ip)
//	if err != nil {
//		return false, err
//	}
//
//}
