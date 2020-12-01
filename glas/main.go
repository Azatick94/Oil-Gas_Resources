// glas
// Copyright 2018 softlandia@gmail.com
// Обработка las файлов. Построение словаря и замена мнемоник на справочные

// TODO может быть сделать возможность задавать правила на которые выполняется проверка
//      например: FIELD == "Весеннее"
//      соответственно при команде INFO в журнал выводятся случаи нарушения данного правила

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/softlandia/xlib"
	"gopkg.in/ini.v1"
)

const (
	fileNameMnemonic = "mnemonic.ini"
	globalConfigName = "glas.ini"
	configFileName   = "las.ini"
)

var (
	//Cfg - global programm config
	Cfg *Config
	//Mnemonic - map of std mnemonic
	Mnemonic map[string]string
	//Dic - mnemonic substitution dictionary
	Dic map[string]string
)

// comandLineParameters - check input parameters
// >glas w "d:\input" "e:\output"
func comandLineParameters() bool {
	//back door for TEST
	if (len(os.Args) == 2) && (os.Args[1] == "-") {
		color.Set(color.FgYellow, color.Bold)
		fmt.Printf("WARNING! using glas.ini command\n")
		color.Unset()
		return true
	}
	if len(os.Args) < 3 { //minimum: glas w d:\input
		fmt.Printf("using: >glas c 'd:\\input' 'e:\\output'\n")
		fmt.Printf("c - command: i, x\n")
		fmt.Printf("'d:\\input'  - path to exist folder\n")
		fmt.Printf("'d:\\output' - path to exist folder\n")
		return false
	}
	if len(os.Args[1]) > 1 { //command only one symbol
		fmt.Print("using: >glas ")
		color.Set(color.FgYellow, color.Bold)
		fmt.Print("C")
		color.Unset()
		fmt.Printf(" 'd:\\input' 'e:\\output'\n")
		fmt.Printf("C - command must be 'i' or 'x'\n")
		fmt.Printf("You entered the wrong command: '%s'\n", os.Args[1])
		return false
	}
	switch os.Args[1] {
	case "i":
		Cfg.Comand = "info"
	case "x":
		Cfg.Comand = "repair"
	default:
		Cfg.Comand = "test"
	}
	if !xlib.FileExists(os.Args[2]) {
		fmt.Print("using: >glas w 'd:\\input' 'e:\\output'\n")
		fmt.Print("input folder: ")
		color.Set(color.FgYellow, color.Bold)
		fmt.Printf("'%s'", os.Args[2])
		color.Unset()
		fmt.Print(" not exist\n")
		return false
	}
	//path to input folder ok
	Cfg.Path = os.Args[2]

	if len(os.Args) == 4 { //full version: glas x c:\idata c:\odata
		if !xlib.FileExists(os.Args[3]) {
			fmt.Print("using: >glas w 'd:\\input' ")
			fmt.Printf("'e:\\output'\n")
			fmt.Printf("output folder: ")
			color.Set(color.FgYellow, color.Bold)
			fmt.Printf("'%s'", os.Args[3])
			color.Unset()
			fmt.Printf(" not exist\n")
			return false
		}
		//path to otput folder ok
		Cfg.pathToRepaire = os.Args[3]
	}
	return true
}

//============================================================================
func main() {
	log.Println("start ", os.Args[0])
	//configuration & dictionaries are filled in here
	//initialize() stop programm if error occure
	//initialize() read glas.ini file and configure global var Cfg
	initialize()
	//comand line parameters rather then ini file
	//and redefine if exist
	if !comandLineParameters() {
		os.Exit(1)
	}
	color.Set(color.FgCyan)
	fmt.Printf("init:\tok.\n")
	fmt.Printf("precision:\t%v\n", Cfg.Epsilon)
	fmt.Printf("debug level:\t%v\n", Cfg.LogLevel)
	fmt.Printf("dictionary:\t%v\n", Cfg.DicFile)
	fmt.Printf("input path:\t%v\n", Cfg.Path)
	fmt.Printf("output path:\t%v\n", Cfg.pathToRepaire)
	fmt.Printf("std NULL:\t%v\n", Cfg.Null)
	fmt.Printf("replace NULL:\t%v\n", Cfg.NullReplace)
	fmt.Printf("verify date:\t%v\n", Cfg.verifyDate)
	fmt.Printf("report files:\t'%s', '%s'\n", Cfg.logFailReport, Cfg.logGoodReport)
	fmt.Printf("message report:\t%s\n", Cfg.lasMessageReport)
	fmt.Printf("missing log:\t%s\n", Cfg.logMissingReport)
	fmt.Printf("warning report:\t%s\n", Cfg.lasWarningReport)

	color.Set(color.FgYellow, color.Bold)
	fmt.Printf("command:\t%v\n", Cfg.Comand)
	color.Unset()

	fileList := make([]string, 0, 10)
	//makeFilesList() stop programm if error occure
	n := makeFilesList(&fileList, Cfg.Path)

	switch Cfg.Comand {
	case "test":
		TEST(n)
	case "convert":
		log.Println("convert code page: ")
		convertCodePage(&fileList)
	case "verify":
		log.Println("verify las:")
	case "repair":
		log.Println("repaire las:")
		repairLas(&fileList, &Dic, Cfg.Path, Cfg.pathToRepaire, Cfg.lasMessageReport, Cfg.lasWarningReport)
	case "info":
		log.Println("collect log info:")
		statisticLas(&fileList, &Dic, Cfg.logFailReport, Cfg.lasMessageReport, Cfg.lasWarningReport, Cfg.logMissingReport)
	}
}

///////////////////////////////////////////
func verifyLas(fl *[]string) error {
	log.Printf("action 'verify' not define")
	return nil
}

///////////////////////////////////////////
func convertCodePage(fl *[]string) error {
	log.Printf("action 'convert' not define")
	return nil
}

////////////////////////////////////////////////////////////
//load std mnemonic
func readGlobalMnemonic(iniFileName string) (map[string]string, error) {
	iniMnemonic, err := ini.Load(iniFileName)
	if err != nil {
		log.Printf("error on load std mnemonic, check out file 'mnemonic.ini'\n")
		return nil, err
	}
	sec, err := iniMnemonic.GetSection("mnemonic")
	if err != nil {
		log.Printf("error on read 'mnemonic.ini'")
		return nil, err
	}
	x := make(map[string]string)
	for _, s := range sec.KeyStrings() {
		x[s] = sec.Key(s).Value()
	}
	if Cfg.LogLevel == "DEBUG" {
		log.Println("__mnemonics:")
		for k, v := range x {
			fmt.Printf("mnemonic: %s, desc: %s\n", k, v)
		}
	}
	return x, nil
}

////////////////////////////////////////////////////////////
//init programm, read config, ini files and init dictionary
//stop programm if not successful
func initialize() {
	var err error

	//read global config from yaml file
	Cfg, err = readGlobalConfig(configFileName)
	if err != nil {
		log.Printf("Fail read '%s' config file. %v", configFileName, err)
		os.Exit(1)
	}

	Mnemonic, err = readGlobalMnemonic(fileNameMnemonic)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	//read dectionary from ini file
	iniDic, err := ini.Load(Cfg.DicFile)
	if err != nil {
		log.Printf("Fail to read file: %v\n", err)
		os.Exit(2)
	}
	sec, err := iniDic.GetSection("LOG")
	if err != nil {
		log.Println("Fail load section 'LOG' from file 'dic.ini'. ", err)
		os.Exit(3)
	}
	//fill dictionary
	Dic = make(map[string]string)
	for _, s := range sec.KeyStrings() {
		Dic[s] = sec.Key(s).Value()
	}
	//словарь заполнен
	if Cfg.LogLevel == "DEBUG" {
		log.Println("__dic:")
		for k, v := range Dic {
			fmt.Println("key: ", k, " val: ", v)
		}
	}
}

// makeFilesList - find and load to array all founded las files
//TODO makeFilesList() there is no need to exit the program when an error occurs
func makeFilesList(fileList *[]string, path string) int {
	n, err := xlib.FindFilesExt(fileList, Cfg.Path, ".las")
	if err != nil {
		log.Println("error at search files. verify path: ", Cfg.Path, err)
		log.Println("stop")
		os.Exit(4)
	}
	if n == 0 {
		log.Println("files 'las' not found. verify parameter path: '", Cfg.Path, "' and change in 'main.yaml'")
		log.Println("stop")
		os.Exit(5)
	}
	if Cfg.LogLevel == "DEBUG" {
		log.Println("founded ", n, " las files:")
		if Cfg.LogLevel == "DEBUG" {
			for i, s := range *fileList {
				log.Println(i, " : ", s)
			}
		}
	}
	return n
}

// TEST - test read and write las files
//TODO report las.warning.md written without filename
func TEST(m int) {
	//test file "1.las"
	las := NewLas()
	n, err := las.Open("1.las")
	if n == 7 {
		fmt.Println("TEST read 1.las OK")
		fmt.Println(err)
	} else {
		fmt.Printf("TEST read 1.las ERROR, n = %d, must 7\n", n)
		fmt.Println(err)
	}

	err = las.setNull(Cfg.Null)
	fmt.Println("set new null value done, error: ", err)

	err = las.Save("-1.las")
	if err != nil {
		fmt.Println("TEST save -1.las ERROR: ", err)
	} else {
		fmt.Println("TEST save -1.las OK")
	}

	las = nil
	las = NewLas()
	n, err = las.Open("-1.las")
	if (n == 7) && (las.Null == -999.25) {
		fmt.Println("TEST read -1.las OK")
		fmt.Println(err)
	} else {
		fmt.Println("TEST read -1.las ERROR")
		fmt.Println("NULL not -999.25 or count dept points != 7")
		fmt.Println(err)
	}

	las = nil
	las = NewLas()
	n, err = las.Open("2.las")
	if n == 4895 {
		fmt.Println("TEST read 2.las OK")
		fmt.Println(err)
	} else {
		fmt.Println("TEST read 2.las ERROR")
		fmt.Println(err)
	}
	err = las.Save("-2.las")
	if err != nil {
		fmt.Println("TEST save -2.las ERROR")
		fmt.Println(err)
	} else {
		fmt.Println("TEST save -2.las OK")
	}
	las = nil

	las = NewLas()
	n, err = las.Open("4.las")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	if n == 23 {
		fmt.Printf("TEST read 4.las OK, count data must 23, actualy: %d\n", n)
	} else {
		fmt.Printf("TEST read 4.las ERROR, count data must 23, actualy: %d\n", n)
	}
	oFile, _ := os.Create(Cfg.lasWarningReport)
	defer oFile.Close()
	oFile.WriteString("file: " + las.FileName + "\n")
	for i, w := range las.warnings {
		fmt.Fprintf(oFile, "%d, dir: %d,\tsec: %d,\tl: %d,\tdesc: %s\n", i, w.direct, w.section, w.line, w.desc)
	}

	las.FileName = "-4.las"
	err = las.Save(las.FileName)
	if err != nil {
		fmt.Printf("error: %v on save file -4.las\n", err)
	} else {
		fmt.Printf("save file -4.las OK\n")
	}
	oFile.WriteString("save file: " + las.FileName + "\n")
	for i, w := range las.warnings {
		fmt.Fprintf(oFile, "%d, dir: %d,\tsec: %d,\tl: %d,\tdesc: %s\n", i, w.direct, w.section, w.line, w.desc)
	}
	las = nil
}
