package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"seckill/common"
	"seckill/datamodels"
	"seckill/encrypt"
	"seckill/rabbitmq"
	"strconv"
	"sync"
	"time"
)

var (
	// 设置集群地址，最好内外IP
	hostArray = []string{"172.20.93.188", "172.20.93.188"}

	localHost = ""

	// 数量控制接口服务器内网IP，或者getone的SLP内网IP
	GetOneIP = "127.0.0.1"

	GetOnePort = "8084"

	port = "8083"

	hashConsistent *common.Consistent

	// rabbitmq
	rabbitMQValidate *rabbitmq.RabbitMQ
)

// 用来存放控制信息
type AccessControl struct {
	// 用来存放用户想要存放的信息
	sourcesArray map[int]time.Time
	sync.RWMutex
}

// 服务器间隔时间，单位秒
var interval = 20

// 创建全局变量
var accessControl = &AccessControl{sourcesArray: make(map[int]time.Time)}

// 获取指定的数据
func (m *AccessControl) GetNewRecord(uid int) time.Time {
	m.RWMutex.RLock()
	defer m.RWMutex.RUnlock()
	return m.sourcesArray[uid]
}

// 设置记录
func (m *AccessControl) SetNewRecord(uid int) {
	m.RWMutex.Lock()
	m.sourcesArray[uid] = time.Now()
	m.RWMutex.Unlock()
}

type BlackList struct {
	listArray map[int]bool
	sync.RWMutex
}

var blackList = &BlackList{listArray: make(map[int]bool)}

// 获取黑名单
func (m *BlackList) GetBlackListByID(uid int) bool {
	m.RLock()
	defer m.RUnlock()
	return m.listArray[uid]
}

// 添加黑名单
func (m *BlackList) SetBlackListByID(uid int) bool {
	m.Lock()
	defer m.Unlock()
	m.listArray[uid] = true
	return true
}

func (m *AccessControl) GetDistributedRight(req *http.Request) bool {
	// 获取用户UID
	uid, err := req.Cookie("uid")
	if err != nil {
		return false
	}

	// 采用一致性hash算法，根据用户ID，判断获取具体机器
	hostRequest, err := hashConsistent.Get(uid.Value)
	if err != nil {
		return false
	}

	// 判断是否为本机
	if hostRequest == localHost {
		// 执行本机数据读取和校验
		return m.GetDataFromMap(uid.Value)
	} else {
		// 不是本机充当代理访问数据返回结果
		return GetDataFromOtherMap(hostRequest, req)
	}
}

// 获取本机map，并且处理业务逻辑，返回的结果类型为bool类型
func (m *AccessControl) GetDataFromMap(uid string) (isOk bool) {
	uidInt, err := strconv.Atoi(uid)
	if err != nil {
		return false
	}

	// 添加黑名单
	if blackList.GetBlackListByID(uidInt) {
		// 判断是否被添加到黑名单
		return false
	}

	// 获取记录
	dataRecord := m.GetNewRecord(uidInt)
	if !dataRecord.IsZero() {
		// 业务判断，是否在指定时间之后
		if dataRecord.Add(time.Duration(interval) * time.Second).After(time.Now()) {
			return false
		}
	}
	m.SetNewRecord(uidInt)
	return true
}

// 获取其他节点处理结果
func GetDataFromOtherMap(host string, request *http.Request) bool {
	hostUrl := "http://" + host + ":" + port + "/checkRight"
	response, body, err := GetCurl(hostUrl, request)
	if err != nil {
		return false
	}

	// 判断状态
	if response.StatusCode == 200 {
		if string(body) == "true" {
			return true
		} else {
			return false
		}
	}
	return false
}

// 模拟请求
func GetCurl(hostUrl string, request *http.Request) (response *http.Response, body []byte, err error) {
	// 获取uid
	uidPre, err := request.Cookie("uid")
	if err != nil {
		return
	}

	// 获取sign
	uidSign, err := request.Cookie("sign")
	if err != nil {
		return
	}

	// 模拟接口访问
	client := &http.Client{}
	req, err := http.NewRequest("GET", hostUrl, nil)
	if err != nil {
		return
	}

	// 手动指定，排除多余cookie
	cookieUid := &http.Cookie{Name: "uid", Value: uidPre.Value, Path: "/"}
	cookieSign := &http.Cookie{Name: "sign", Value: uidSign.Value, Path: "/"}

	// 添加cookie到模拟的请求中
	req.AddCookie(cookieUid)
	req.AddCookie(cookieSign)

	// 获取返回结果
	response, err = client.Do(req)
	defer response.Body.Close()
	if err != nil {
		return
	}
	body, err = ioutil.ReadAll(response.Body)
	return
}

func CheckRight(w http.ResponseWriter, r *http.Request) {
	right := accessControl.GetDistributedRight(r)
	if !right {
		w.Write([]byte("false"))
		return
	}
	w.Write([]byte("true"))
	return
}

// 执行正常业务逻辑
func Check(w http.ResponseWriter, r *http.Request) {
	// 执行正常业务逻辑
	fmt.Println("执行check")
	queryForm, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil || len(queryForm["productID"]) <= 0 {
		w.Write([]byte("false"))
		return
	}
	productString := queryForm["productID"][0]
	fmt.Println(productString)

	// 获取用户cookie
	userCookie, err := r.Cookie("uid")
	if err != nil {
		w.Write([]byte("false"))
		return
	}

	// 1.分布式权限验证
	right := accessControl.GetDistributedRight(r)
	if !right {
		w.Write([]byte("false"))
		return
	}

	// 2.获取数量控制权限，防止秒杀出现超买现象
	hostUrl := "http://" + GetOneIP + ":" + GetOnePort + "/getOne"
	responseValidate, validateBody, err := GetCurl(hostUrl, r)
	if err != nil {
		w.Write([]byte("false"))
		return
	}
	// 判断数量控制接口请求状态
	if responseValidate.StatusCode == 200 {
		if string(validateBody) == "true" {
			// 整合下单
			// 1.获取商品ID
			productID, err := strconv.ParseInt(productString, 10, 64)
			if err != nil {
				w.Write([]byte("false"))
				return
			}
			// 2.获取用户ID
			userID, err := strconv.ParseInt(userCookie.Value, 10, 64)
			if err != nil {
				w.Write([]byte("false"))
				return
			}
			// 3.创建消息体
			message := datamodels.NewMessage(userID, productID)
			// 类型转化
			byteMessage, err := json.Marshal(message)
			if err != nil {
				w.Write([]byte("false"))
				return
			}
			// 4.生产消息
			err = rabbitMQValidate.PublishSimple(string(byteMessage))
			if err != nil {
				w.Write([]byte("false"))
				return
			}
			w.Write([]byte("true"))
			return
		}
	}
	w.Write([]byte("false"))
	return
}

// 统一验证拦截器，每一个接口都需要提前验证
func Auth(w http.ResponseWriter, r *http.Request) error {
	fmt.Println("执行验证")

	// 添加基于cookie的权限验证
	err := CheckUserInfo(r)
	if err != nil {
		return err
	}
	return nil
}

// 身份校验函数
func CheckUserInfo(r *http.Request) error {
	// 获取UID，cookie
	uidCookie, err := r.Cookie("uid")
	if err != nil {
		return errors.New("用户UID Cookie获取失败")
	}

	// 获取用户加密串
	signCookie, err := r.Cookie("sign")
	if err != nil {
		return errors.New("用户加密串Cookie获取失败")
	}

	// 对信息进行解密
	signByte, err := encrypt.DePwdCode(signCookie.Value)
	if err != nil {
		return errors.New("加密串已被篡改")
	}

	if checkInfo(uidCookie.Value, string(signByte)) {
		return nil
	}
	return errors.New("身份校验失败")
}

// 自定义逻辑判断
func checkInfo(checkStr, signStr string) bool {
	if checkStr == signStr {
		return true
	}
	return false
}

func main() {
	// 负载均衡器设置，采用一致性哈希算法
	hashConsistent = common.NewConsistent()

	// 采用一致性hash算法，添加节点
	for _, v := range hostArray {
		hashConsistent.Add(v)
	}

	localIP, err := common.GetIntranceIP()
	if err != nil {
		fmt.Println(err)
	}
	localHost = localIP
	fmt.Println(localHost)

	rabbitMQValidate = rabbitmq.NewRabbitMQSimple("product")
	defer rabbitMQValidate.Destory()

	// 设置静态文件目录
	http.Handle("/html/", http.StripPrefix("/html/", http.FileServer(http.Dir("./frontend/web/htmlProductShow"))))

	// 设置资源目录
	http.Handle("/public/", http.StripPrefix("/public", http.FileServer(http.Dir("./frontend/web/public"))))

	// 1.过滤器
	filter := common.NewFilter()
	// 注册拦截器
	filter.RegisterFilterURL("/check", Auth)
	filter.RegisterFilterURL("/checkRight", Auth)

	// 2.启动服务
	http.HandleFunc("/check", filter.Handle(Check))
	http.HandleFunc("/checkRight", filter.Handle(CheckRight))
	// 启动服务
	http.ListenAndServe(":8083", nil)
}
