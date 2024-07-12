package login

import (
	"github.com/eatmoreapple/openwechat"
	"log"
)

func HandleHotLogin(bot *openwechat.Bot) {
	// 创建热存储容器对象
	reloadStorage := openwechat.NewFileHotReloadStorage("storage.json")

	defer reloadStorage.Close()

	// 执行热登录
	if err := bot.HotLogin(reloadStorage, openwechat.NewRetryLoginOption()); err != nil {
		log.Fatalf("HotLogin failed: %v", err)
		// 登录并扫码
		if err := bot.Login(); err != nil {
			log.Fatalf("Login failed: %v", err)
		}
	}
}
