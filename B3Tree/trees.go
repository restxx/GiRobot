package B3Tree

import (
	b3 "github.com/magicsea/behavior3go"
	"github.com/magicsea/behavior3go/config"
	"github.com/magicsea/behavior3go/core"
	"github.com/magicsea/behavior3go/loader"
	logger "github.com/restxx/GiRobot/Logger"
)

var allTrees = make(map[string]*core.BehaviorTree)

func GetTree(title string) *core.BehaviorTree {
	t, ok := allTrees[title]
	if ok {
		return t
	}
	return nil
}

func SetTree(title string, tree *core.BehaviorTree) {
	allTrees[title] = tree
}

var StructMaps = b3.NewRegisterStructMaps()

func Init(fileName string) {
	projectConfig, ok := config.LoadRawProjectCfg(fileName)
	if !ok {
		panic("LoadRawProjectCfg Failed: " + fileName)
	}

	for _, v := range projectConfig.Data.Trees {
		SetTree(v.Title, loader.CreateBevTreeFromConfig(&v, StructMaps))
		logger.Info("[BT] 加载行为树 %s", v.Title)
	}
}
