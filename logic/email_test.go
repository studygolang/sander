package logic_test

import (
	"sander/logic"
	"testing"

	"sander/config"
	"sander/logger"
)

func TestSendMail(t *testing.T) {
	logger.Init(ROOT+"/log", config.ConfigFile.MustValue("global", "log_level", "DEBUG"))

	err := logic.DefaultEmail.SendMail("中文test", "内容test content，收到？", []string{"xuxinhua@zhimadj.com"})
	if err != nil {
		t.Error(err)
	} else {
		t.Log("successful")
	}
}
