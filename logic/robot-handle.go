package logic

import (
	logger "GiantQA/GiRobot/Logger"
	"GiantQA/GiRobot/utils"
	"GiantQA/common"
	csproto "GiantQA/proto"
	"fmt"
)

type HandleMessage struct {
	utils.FuncMap
}

var HandleMsg = &HandleMessage{}

// 获取角色信息
var _ = HandleMsg.Bind(csproto.Cmd_GAME_ROLEINFO, func(r *DBTRobotV2, Msg *csproto.CommonMessage) {
	r.RoleInfo = Msg.GetRoleInfoNtf()
})

// 得到邮件
var _ = HandleMsg.Bind(csproto.Cmd_MAIL_LIST, func(r *DBTRobotV2, Msg *csproto.CommonMessage) {
	if r.MailInfo == nil {
		r.MailInfo = Msg.GetMailInfoNtf()
		return
	}
	r.MailInfo.Mails = append(r.MailInfo.Mails, Msg.GetMailInfoNtf().Mails...)
	fmt.Println(r.MailInfo)
})

// 获取英雄信息
var _ = HandleMsg.Bind(csproto.Cmd_GAME_CHAR, func(r *DBTRobotV2, Msg *csproto.CommonMessage) {
	r.CharInfo = Msg.GetRoleCharNtf()
})

// Package信息
var _ = HandleMsg.Bind(csproto.Cmd_GAME_PACKAGE, func(r *DBTRobotV2, Msg *csproto.CommonMessage) {
	pkg := Msg.GetRolePackageNtf()
	if pkg != nil {
		cur := pkg.GetEquips()
		if cur != nil {
			r.Equips = append(r.Equips, cur...)
		}
		skills := pkg.GetSkills()
		if skills != nil {
			for _, skill := range skills {
				if skill.Package != -1 && skill.Count > 0 { // 在身上的不能升级
					r.Skills = append(r.Skills, skill)
				}
			}
		}
	}
})

// 公共消息 更新客户端数据
var _ = HandleMsg.Bind(csproto.Cmd_GAME_GAIN_NTF, func(r *DBTRobotV2, Msg *csproto.CommonMessage) {

	defer func() {
		if p := recover(); p != nil {
			logger.Error("[%s] Cmd_GAME_GAIN_NTF %v,  r.CharInfo = %v  r.MailInfo = %v", r.Account(), Msg.GetGainInfoNtf(), r.CharInfo, r.MailInfo)
		}
	}()
	logger.Info("[%s] Cmd_GAME_GAIN_NTF %v", r.Account(), Msg.GetGainInfoNtf())
	msg := Msg.GetGainInfoNtf()

	for _, v := range msg.Datas {
		if v.Dt == csproto.GainType_EQUIP {
			if len(v.Equips) == 0 {
				continue
			}
			if v.Equips[0].Package == -1 { //穿
				r.Equips = nil
				r.CharInfo.Chars[0].Equips = make([]*csproto.EquipInfo, 0)
			} else {
				r.Equips = v.Equips
				r.CharInfo.Chars[0].Equips = nil
			}
		} else if v.Dt == csproto.GainType_MAIL {
			logger.Info("[%s]收到邮件Gain消息 mailId=[%v]", r.Account(), v.Mail)
			b, body := common.Base64Decode([]byte(v.Mail.Body))
			if !b {
				logger.Error("[%s]base64 解body错误", r.Account())
				continue
			}
			v.Mail.Body = string(body)
			if r.MailInfo == nil {
				r.MailInfo = &csproto.MailInfoNtf{}
			}
			r.MailInfo.Mails = append(r.MailInfo.Mails, v.Mail)
		}
	}
})

// 测试消息
var _ = HandleMsg.Bind(csproto.Cmd_WORLD_CHAT_NTF, func(r *DBTRobotV2, Msg *csproto.CommonMessage) {

	for _, v := range Msg.WorldChatNtf.Msgs {
		if v.SrcRole.RoleId == r.RoleInfo.RoleId {
			logger.Debug("[%s] ------ Chat", r.Account())
		}
	}
})
