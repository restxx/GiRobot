package utils

import (
	"bytes"
	"encoding/binary"
	"io"
)

// 定义服务器消息类型信息,类型取值范围 [0, 15],如果未指定value属性,则由程序采用自增方式自动生成
const (
	LOGINSERVER uint32 = 0x6 // 登录服务器的消息类型
	WORLDSERVER uint32 = 0x5 // 游戏服务器消息类型
)

// 定义服务器消息类型ID,如果未指定value属性,则有程序采用自增方式自动生成
const (
	Version uint16 = 11
	// 游戏大厅消息,定义登录服务器消息
	FailReason            uint32 = LOGINSERVER<<28 | 0x000c3501 // 游戏服务器通用错误结构体
	RequestLogin          uint32 = LOGINSERVER<<28 | 0x000c8321 // 客户端请求登录
	ResponseLogin         uint32 = LOGINSERVER<<28 | 0x000c8322 // 登录服务器回馈登录请求，发送令牌
	RequestAccountInfo    uint32 = LOGINSERVER<<28 | 0x000c8323 // 请求数据库查询账户信息
	ResponseAccountInfo   uint32 = LOGINSERVER<<28 | 0x000c8324 // 账号信息数据库
	NotifyGameServiceInfo uint32 = LOGINSERVER<<28 | 0x000c8325 // 通知客户端游戏服务器列表
	NotifyHeartbeat       uint32 = LOGINSERVER<<28 | 0x000dbba0 // 游戏服务器活跃心跳

	RequestEnterWorld  uint32 = WORLDSERVER<<28 | 0x0001adb0 // 客户端进入游戏服务器
	ResponseEnterWorld uint32 = WORLDSERVER<<28 | 0x0001adb1 // 应答客户端进入游戏服务器事件

	RequestBuyGoods  uint32 = WORLDSERVER<<28 | 0x0002f1fa // 请求购买商品
	ResponseBuyGoods uint32 = WORLDSERVER<<28 | 0x0002f1fb // 返回购买结果

)

type IMessage interface {
	ToBytes() []byte
	FromReader(buf io.Reader)
}

// 客户端请求登录服务器
type CRequestLogin struct {
	// CSerializable
	M_nRuntimeTypeId uint32
	M_nTargetId_base uint16
	M_nSourceId      uint16
	M_nVersion       uint16

	M_nAccountLength byte
	M_szAccount      []byte // 用户名或者设备ID
	M_nTokenLength   uint16
	M_szToken        []byte // 客户端密钥数据
	M_nLoginMode     byte   // 登录方式,0标识采用帐号密码登录,否则为设备登录
}

var _ = MessageMap.Bind(RequestLogin, func() interface{} {
	return &CRequestLogin{M_nRuntimeTypeId: RequestLogin, M_nVersion: Version}
})

func (this *CRequestLogin) ToBytes() []byte {
	buf := bytes.NewBuffer([]byte{})
	var err error

	// CSerializable
	err = binary.Write(buf, Endian, this.M_nRuntimeTypeId)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nTargetId_base)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nSourceId)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nVersion)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nAccountLength)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_szAccount)
	if err != nil {
		panic(err)
	}
	err = binary.Write(buf, Endian, this.M_nTokenLength)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_szToken)
	if err != nil {
		panic(err)
	}
	err = binary.Write(buf, Endian, this.M_nLoginMode)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func (this *CRequestLogin) FromReader(buf io.Reader) {
	var err error

	// CSerializable
	err = binary.Read(buf, Endian, &this.M_nRuntimeTypeId)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nTargetId_base)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nSourceId)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nVersion)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nAccountLength)
	if err != nil {
		panic(err)
	}

	this.M_szAccount = make([]byte, this.M_nAccountLength)
	err = binary.Read(buf, Endian, this.M_szAccount)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nTokenLength)
	if err != nil {
		panic(err)
	}

	this.M_szToken = make([]byte, this.M_nTokenLength)
	err = binary.Read(buf, Endian, this.M_szToken)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nLoginMode)
	if err != nil {
		panic(err)
	}

}

// 登录服务器回馈登录结果
type CResponseLogin struct {
	// CSerializable
	M_nRuntimeTypeId uint32
	M_nTargetId_base uint16
	M_nSourceId      uint16
	M_nVersion       uint16

	// CFaileReason
	M_nResult int16

	M_nAccountId        uint32 // 系统唯一帐号标识ID
	M_nTokenLength      uint16
	M_szToken           []byte // 身份验证令牌
	M_nSessionKeyLength byte
	M_szSessionKey      []byte // 服务器通信回话密钥
	M_nServiceId        uint16 // 服务器建议游戏服务器ID
	M_nOpenTick         uint32 // 服务器开启时间
}

var _ = MessageMap.Bind(ResponseLogin, func() interface{} {
	return &CResponseLogin{M_nRuntimeTypeId: ResponseLogin, M_nVersion: Version}
})

func (this *CResponseLogin) ToBytes() []byte {
	buf := bytes.NewBuffer([]byte{})
	var err error

	// CSerializable
	err = binary.Write(buf, Endian, this.M_nRuntimeTypeId)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nTargetId_base)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nSourceId)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nVersion)
	if err != nil {
		panic(err)
	}

	// CFailReason
	err = binary.Write(buf, Endian, this.M_nResult)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nAccountId)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nTokenLength)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_szToken)
	if err != nil {
		panic(err)
	}
	err = binary.Write(buf, Endian, this.M_nSessionKeyLength)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_szSessionKey)
	if err != nil {
		panic(err)
	}
	err = binary.Write(buf, Endian, this.M_nServiceId)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nOpenTick)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func (this *CResponseLogin) FromReader(buf io.Reader) {
	var err error

	// CSerializable
	err = binary.Read(buf, Endian, &this.M_nRuntimeTypeId)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nTargetId_base)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nSourceId)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nVersion)
	if err != nil {
		panic(err)
	}

	// CFailReason
	err = binary.Read(buf, Endian, &this.M_nResult)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nAccountId)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nTokenLength)
	if err != nil {
		panic(err)
	}

	this.M_szToken = make([]byte, this.M_nTokenLength)
	err = binary.Read(buf, Endian, this.M_szToken)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nSessionKeyLength)
	if err != nil {
		panic(err)
	}

	this.M_szSessionKey = make([]byte, this.M_nSessionKeyLength)
	err = binary.Read(buf, Endian, this.M_szSessionKey)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nServiceId)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nOpenTick)
	if err != nil {
		panic(err)
	}

}

// 网络地址
type CServiceInfo struct {
	M_nServiceId  uint16 // 游戏服务器标识ID
	M_nNameLength byte   // 主机名称大小
	M_szHostName  []byte // 游戏服务器名称
	M_nLength     byte   // 主机名称大小
	M_szHost      []byte // IP地址
	M_nPort       uint16 // 游戏服务器监听端口
	M_nLoad       byte   // 服务器当前负载
}

func (this *CServiceInfo) ToBytes() []byte {
	buf := bytes.NewBuffer([]byte{})
	var err error

	err = binary.Write(buf, Endian, this.M_nServiceId)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nNameLength)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_szHostName)
	if err != nil {
		panic(err)
	}
	err = binary.Write(buf, Endian, this.M_nLength)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_szHost)
	if err != nil {
		panic(err)
	}
	err = binary.Write(buf, Endian, this.M_nPort)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nLoad)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func (this *CServiceInfo) FromReader(buf io.Reader) {
	var err error

	err = binary.Read(buf, Endian, &this.M_nServiceId)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nNameLength)
	if err != nil {
		panic(err)
	}

	this.M_szHostName = make([]byte, this.M_nNameLength)
	err = binary.Read(buf, Endian, this.M_szHostName)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nLength)
	if err != nil {
		panic(err)
	}

	this.M_szHost = make([]byte, this.M_nLength)
	err = binary.Read(buf, Endian, this.M_szHost)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nPort)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nLoad)
	if err != nil {
		panic(err)
	}

}

// 游戏服务器地址信息
type CNotifyGameServiceInfo struct {
	// CSerializable
	M_nRuntimeTypeId uint32
	M_nTargetId_base uint16
	M_nSourceId      uint16
	M_nVersion       uint16

	M_nLength      byte
	M_nServiceInfo []CServiceInfo // 用户帐号名称
}

var _ = MessageMap.Bind(NotifyGameServiceInfo, func() interface{} {
	return &CNotifyGameServiceInfo{M_nRuntimeTypeId: NotifyGameServiceInfo, M_nVersion: Version}
})

func (this *CNotifyGameServiceInfo) ToBytes() []byte {
	buf := bytes.NewBuffer([]byte{})
	var err error

	// CSerializable
	err = binary.Write(buf, Endian, this.M_nRuntimeTypeId)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nTargetId_base)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nSourceId)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nVersion)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nLength)
	if err != nil {
		panic(err)
	}

	for i := 0; i < int(this.M_nLength); i++ {
		_, err = buf.Write(this.M_nServiceInfo[i].ToBytes())
		if err != nil {
		}
	}

	return buf.Bytes()
}

func (this *CNotifyGameServiceInfo) FromReader(buf io.Reader) {
	var err error

	// CSerializable
	err = binary.Read(buf, Endian, &this.M_nRuntimeTypeId)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nTargetId_base)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nSourceId)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nVersion)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nLength)
	if err != nil {
		panic(err)
	}

	this.M_nServiceInfo = make([]CServiceInfo, this.M_nLength)
	for i := 0; i < int(this.M_nLength); i++ {
		this.M_nServiceInfo[i].FromReader(buf)
	}
}

// 玩家基础信息
type CRoleInfo struct {
	M_nLength     byte   // 昵称长度
	M_szNick      []byte // 昵称
	M_nSignLength byte   // 简介长度
	M_szSign      []byte // 个人简介
	M_nPictureId  uint32 // 玩家头像ID
	M_nBuyGold    uint32 // 玩家消费统计,用于计算VIP等级
	M_nCostGold   uint32 // 玩家消费统计
	M_nExp        uint32 // 玩家游戏经验
	M_nEnergy     uint32 // 体力值
	M_nEnergyTick uint32 // 体力恢复时间戳
	M_nBagExt     uint32 // 背包扩展
	M_nGrowth     uint32 // 成长记录,新手引导
	M_nPrivilege  uint32 // 游戏服务器特权信息
	M_nFakeVIP    uint16 // 游戏调试VIP值
	M_nAvatarId   uint32 // 上阵英雄
	M_nFashion    uint32 // 英雄时装
	M_nPetId      uint32 // 上阵配宠
	M_nOutfit     uint32 // 上阵方案
}

func (this *CRoleInfo) ToBytes() []byte {
	buf := bytes.NewBuffer([]byte{})
	var err error

	err = binary.Write(buf, Endian, this.M_nLength)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_szNick)
	if err != nil {
		panic(err)
	}
	err = binary.Write(buf, Endian, this.M_nSignLength)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_szSign)
	if err != nil {
		panic(err)
	}
	err = binary.Write(buf, Endian, this.M_nPictureId)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nBuyGold)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nCostGold)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nExp)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nEnergy)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nEnergyTick)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nBagExt)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nGrowth)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nPrivilege)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nFakeVIP)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nAvatarId)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nFashion)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nPetId)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nOutfit)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func (this *CRoleInfo) FromReader(buf io.Reader) {
	var err error

	err = binary.Read(buf, Endian, &this.M_nLength)
	if err != nil {
		panic(err)
	}

	this.M_szNick = make([]byte, this.M_nLength)
	err = binary.Read(buf, Endian, this.M_szNick)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nSignLength)
	if err != nil {
		panic(err)
	}

	this.M_szSign = make([]byte, this.M_nSignLength)
	err = binary.Read(buf, Endian, this.M_szSign)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nPictureId)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nBuyGold)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nCostGold)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nExp)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nEnergy)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nEnergyTick)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nBagExt)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nGrowth)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nPrivilege)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nFakeVIP)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nAvatarId)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nFashion)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nPetId)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nOutfit)
	if err != nil {
		panic(err)
	}
}

// 客户端请求进入游戏世界
type CRequestEnterWorld struct {
	// CSerializable
	M_nRuntimeTypeId uint32
	M_nTargetId_base uint16
	M_nSourceId      uint16
	M_nVersion       uint16

	M_nAccountId   uint32 // 系统唯一帐号标识ID
	M_nTokenLength uint16
	M_szToken      []byte // 身份令牌
	M_nInvite      byte   // 0普通进入，1邀请进入，2断线重连
}

var _ = MessageMap.Bind(RequestEnterWorld, func() interface{} {
	return &CRequestEnterWorld{M_nRuntimeTypeId: RequestEnterWorld, M_nVersion: Version}
})

func (this *CRequestEnterWorld) ToBytes() []byte {
	buf := bytes.NewBuffer([]byte{})
	var err error

	// CSerializable
	err = binary.Write(buf, Endian, this.M_nRuntimeTypeId)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nTargetId_base)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nSourceId)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nVersion)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nAccountId)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nTokenLength)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_szToken)
	if err != nil {
		panic(err)
	}
	err = binary.Write(buf, Endian, this.M_nInvite)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func (this *CRequestEnterWorld) FromReader(buf io.Reader) {
	var err error

	// CSerializable
	err = binary.Read(buf, Endian, &this.M_nRuntimeTypeId)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nTargetId_base)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nSourceId)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nVersion)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nAccountId)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nTokenLength)
	if err != nil {
		panic(err)
	}

	this.M_szToken = make([]byte, this.M_nTokenLength)
	err = binary.Read(buf, Endian, this.M_szToken)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nInvite)
	if err != nil {
		panic(err)
	}

}

// 服务器应答客户端进入游戏世界
type CResponseEnterWorld struct {
	// CSerializable
	M_nRuntimeTypeId uint32
	M_nTargetId_base uint16
	M_nSourceId      uint16
	M_nVersion       uint16

	// CFaileReason
	M_nResult int16

	M_nServerTime uint64    // 服务器时间戳
	M_nRoleInfo   CRoleInfo // 角色信息
	M_nPlayerId   uint16    // 服务器分配客户端位移标识
}

var _ = MessageMap.Bind(ResponseEnterWorld, func() interface{} {
	return &CResponseEnterWorld{M_nRuntimeTypeId: ResponseEnterWorld, M_nVersion: Version}
})

func (this *CResponseEnterWorld) ToBytes() []byte {
	buf := bytes.NewBuffer([]byte{})
	var err error

	// CSerializable
	err = binary.Write(buf, Endian, this.M_nRuntimeTypeId)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nTargetId_base)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nSourceId)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nVersion)
	if err != nil {
		panic(err)
	}

	// CFailReason
	err = binary.Write(buf, Endian, this.M_nResult)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nServerTime)
	if err != nil {
		panic(err)
	}

	_, err = buf.Write(this.M_nRoleInfo.ToBytes())
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nPlayerId)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func (this *CResponseEnterWorld) FromReader(buf io.Reader) {
	var err error

	// CSerializable
	err = binary.Read(buf, Endian, &this.M_nRuntimeTypeId)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nTargetId_base)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nSourceId)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nVersion)
	if err != nil {
		panic(err)
	}

	// CFailReason
	err = binary.Read(buf, Endian, &this.M_nResult)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nServerTime)
	if err != nil {
		panic(err)
	}

	this.M_nRoleInfo.FromReader(buf)
	err = binary.Read(buf, Endian, &this.M_nPlayerId)
	if err != nil {
		panic(err)
	}

}

// 请求购买商品
type CRequestBuyGoods struct {
	// CSerializable
	M_nRuntimeTypeId uint32
	M_nTargetId_base uint16
	M_nSourceId      uint16
	M_nVersion       uint16

	M_nAccountId uint32 // 账户ID
	M_bShopId    byte   // 商店类型
	M_nGoodsId   uint32 // 商品ID
	M_nNum       uint32 // 购买数量
}

var _ = MessageMap.Bind(RequestBuyGoods, func() interface{} {
	return &CRequestBuyGoods{M_nRuntimeTypeId: RequestBuyGoods, M_nVersion: Version}
})

func (this *CRequestBuyGoods) ToBytes() []byte {
	buf := bytes.NewBuffer([]byte{})
	var err error

	// CSerializable
	err = binary.Write(buf, Endian, this.M_nRuntimeTypeId)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nTargetId_base)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nSourceId)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nVersion)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nAccountId)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_bShopId)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nGoodsId)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nNum)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func (this *CRequestBuyGoods) FromReader(buf io.Reader) {
	var err error

	// CSerializable
	err = binary.Read(buf, Endian, &this.M_nRuntimeTypeId)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nTargetId_base)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nSourceId)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nVersion)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nAccountId)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_bShopId)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nGoodsId)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nNum)
	if err != nil {
		panic(err)
	}

}

// 返回购买结果
type CResponseBuyGoods struct {
	// CSerializable
	M_nRuntimeTypeId uint32
	M_nTargetId_base uint16
	M_nSourceId      uint16
	M_nVersion       uint16

	// CFaileReason
	M_nResult int16

	M_bShopId     byte   // 商店类型
	M_nGoodsId    uint32 // 商品ID
	M_nSuccessNum uint32 // 购买成功数量
	M_nCostId     uint32 // 消耗道具
	M_nCostCnt    uint32 // 消耗数量
}

var _ = MessageMap.Bind(ResponseBuyGoods, func() interface{} {
	return &CResponseBuyGoods{M_nRuntimeTypeId: ResponseBuyGoods, M_nVersion: Version}
})

func (this *CResponseBuyGoods) ToBytes() []byte {
	buf := bytes.NewBuffer([]byte{})
	var err error

	// CSerializable
	err = binary.Write(buf, Endian, this.M_nRuntimeTypeId)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nTargetId_base)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nSourceId)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nVersion)
	if err != nil {
		panic(err)
	}

	// CFailReason
	err = binary.Write(buf, Endian, this.M_nResult)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_bShopId)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nGoodsId)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nSuccessNum)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nCostId)
	if err != nil {
		panic(err)
	}

	err = binary.Write(buf, Endian, this.M_nCostCnt)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func (this *CResponseBuyGoods) FromReader(buf io.Reader) {
	var err error

	// CSerializable
	err = binary.Read(buf, Endian, &this.M_nRuntimeTypeId)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nTargetId_base)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nSourceId)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nVersion)
	if err != nil {
		panic(err)
	}

	// CFailReason
	err = binary.Read(buf, Endian, &this.M_nResult)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_bShopId)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nGoodsId)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nSuccessNum)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nCostId)
	if err != nil {
		panic(err)
	}

	err = binary.Read(buf, Endian, &this.M_nCostCnt)
	if err != nil {
		panic(err)
	}

}
