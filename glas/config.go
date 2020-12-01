//config.go
package main

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/softlandia/xlib"
	ini "gopkg.in/ini.v1"
	"gopkg.in/yaml.v2"
)

var (
	configModtime  int64
	errNotModified = errors.New("Not modified")
)

// Config - структура для считывания конфигурационного файла
type Config struct {
	//Epsilon - точность обратоки чисел с плавающей запятой
	Epsilon          float64
	LogLevel         string
	DicFile          string
	Path             string
	pathToRepaire    string
	Comand           string
	Null             float64
	NullReplace      bool
	verifyDate       bool
	logGoodReport    string
	logFailReport    string
	logMissingReport string
	lasMessageReport string
	lasWarningReport string
	maxWarningCount  int
}

////////////////////////////////////////////////////////////
func readGlobalConfig(fileName string) (x *Config, err error) {
	x = new(Config)
	gini, err := ini.Load(globalConfigName)
	if err != nil {
		return nil, err
	}
	x.LogLevel = gini.Section("global").Key("loglevel").String()
	x.Epsilon, err = gini.Section("global").Key("epsilon").Float64()
	x.DicFile = gini.Section("global").Key("filedictionary").String()
	x.Path = gini.Section("global").Key("path").String()
	x.pathToRepaire = gini.Section("global").Key("pathToRepaire").String()
	x.Comand = gini.Section("global").Key("cmd").String()
	x.Null, err = gini.Section("global").Key("stdNull").Float64()
	x.NullReplace, err = gini.Section("global").Key("replaceNull").Bool()
	x.verifyDate, err = gini.Section("global").Key("verifyDate").Bool()
	x.logGoodReport = gini.Section("global").Key("logGoodReport").String()
	x.logFailReport = gini.Section("global").Key("logFailReport").String()
	x.logMissingReport = gini.Section("global").Key("logMissingReport").String()
	x.lasMessageReport = gini.Section("global").Key("lasMessageReport").String()
	x.lasWarningReport = gini.Section("global").Key("lasWarningReport").String()
	x.maxWarningCount, _ = gini.Section("global").Key("maxWarningCount").Int()
	return x, err
}

func readConfig(ConfigName string) (x *Config, err error) {
	var file []byte
	if file, err = ioutil.ReadFile(ConfigName); err != nil {
		return nil, err
	}
	x = new(Config)
	if err = yaml.Unmarshal(file, x); err != nil {
		return nil, err
	}
	if x.LogLevel == "" {
		x.LogLevel = "ERROR"
	}
	if x.Epsilon == 0.0 {
		x.Epsilon = xlib.Epsilon
	}
	return x, nil
}

//Проверяет время изменения конфигурационного файла
//и перезагружает его если он изменился
//Возвращает errNotModified если изменений нет
func reloadConfig(configName string) (cfg *Config, err error) {
	info, err := os.Stat(configName)
	if err != nil {
		return nil, err
	}
	if configModtime != info.ModTime().UnixNano() {
		configModtime = info.ModTime().UnixNano()
		cfg, err = readConfig(configName)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}
	return nil, errNotModified
}
