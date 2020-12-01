//
package main

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/softlandia/xlib"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

///format strings represent structure of LAS file
const (
	_LasFirstLine      = "~VERSION INFORMATION\n"
	_LasVersion        = "VERS.                          %3.1f :glas (c) softlandia@gmail.com\n"
	_LasCodePage       = "CPAGE.                         1251: code page \n"
	_LasWrap           = "WRAP.                          NO  : ONE LINE PER DEPTH STEP\n"
	_LasWellInfoSec    = "~WELL INFORMATION\n"
	_LasMnemonicFormat = "#MNEM.UNIT DATA                                  :DESCRIPTION\n"
	_LasStrt           = " STRT.M %8.3f                                    :START DEPTH\n"
	_LasStop           = " STOP.M %8.3f                                    :STOP  DEPTH\n"
	_LasStep           = " STEP.M %8.3f                                    :STEP\n"
	_LasNull           = " NULL.  %9.3f                                   :NULL VALUE\n"
	_LasRkb            = " RKB.M %8.3f                                     :KB or GL\n"
	_LasXcoord         = " XWELL.M %8.3f                                   :Well head X coordinate\n"
	_LasYcoord         = " YWELL.M %8.3f                                   :Well head Y coordinate\n"
	_LasOilComp        = " COMP.  %-43.43s:OIL COMPANY\n"
	_LasWell           = " WELL.   %-43.43s:WELL\n"
	_LasField          = " FLD .  %-43.43s:FIELD\n"
	_LasLoc            = " LOC .  %-43.43s:LOCATION\n"
	_LasCountry        = " CTRY.  %-43.43s:COUNTRY\n"
	_LasServiceComp    = " SRVC.  %-43.43s:SERVICE COMPANY\n"
	_LasDate           = " DATE.  %-43.43s:DATE\n"
	_LasAPI            = " API .  %-43.43s:API NUMBER\n"
	_LasUwi            = " UWI .  %-43.43s:UNIVERSAL WELL INDEX\n"
	_LasCurvSec        = "~Curve Information Section\n"
	_LasCurvFormat     = "#MNEM.UNIT                 :DESCRIPTION\n"
	_LasCurvDept       = " DEPT.M                    :\n"
	_LasCurvLine       = " %s.%s                     :\n"
	_LasCurvLine2      = " %s                        :\n"
	_LasDataSec        = "~A"
	_LasDataLine       = ""

	//secName: 0 - empty, 1 - Version, 2 - Well info, 3 - Curve info, 4 - dAta
	lasSecIgnore   = 0
	lasSecVertion  = 1
	lasSecWellInfo = 2
	lasSecCurInfo  = 3
	lasSecData     = 4
)

//LasLog - class to store one log in Las
type LasLog struct {
	LasParam
	Index int
	dept  []float64
	log   []float64
}

// Las - class to store las file
// input code page autodetect
// at read file always code page converted to UTF
// at save file code page converted to specifyed in Las.toCodePage
//TODO add pointer to cfg
//TODO warnings - need method to flush slice on file, and clear
type Las struct {
	FileName     string             //file name from load
	Ver          float64            //version 1.0, 1.2 or 2.0
	Wrap         string             //YES || NO
	CodePage     string             //1251 //пока не читается
	Strt         float64            //start depth
	Stop         float64            //stop depth
	Step         float64            //depth step
	Null         float64            //value interpreted as empty
	Well         string             //well name
	Rkb          float64            //altitude KB
	Logs         map[string]LasLog  //store all logs
	LogDic       *map[string]string //external dictionary of standart log name - mnemonics
	VocDic       *map[string]string //external vocabulary dictionary of log mnemonic
	expPoints    int                //expected count (.)
	nPoints      int                //actually count (.)
	fromCodePage int                //codepage input file. autodetect
	toCodePage   int                //codepage to save file, default xlib.CpWindows1251. to special value, specify at make: NewLas(cp...)
	iDuplicate   int                //индекс повторящейся мнемоники, увеличивается на 1 при нахождении дубля, начально 0
	currentLine  int                //index of current line in readed file
	warnings     []TWarning         //slice of warnings occure on read or write
	countWarning int                //count of warning, if count > 20 then stop collecting
}

//getStepFromData - return step from data section
//open and read 2 line from section ~A and determine step
//close file
//return <0 if error occure
func (o *Las) getStepFromData(fileName string) float64 {
	iFile, err := os.Open(fileName) // open file to READ
	if err != nil {
		return -1.0
	}
	defer iFile.Close()

	_, iScanner, err := xlib.SeekFileToString(fileName, "~A")
	if err != nil {
		return -1.0
	}

	s := ""
	j := 0
	dept1 := 0.0
	dept2 := 0.0
	for i := 0; iScanner.Scan(); i++ {
		s = strings.TrimSpace(iScanner.Text())
		if (len(s) == 0) || (s[0] == '#') {
			continue
		}
		k := strings.IndexRune(s, ' ')
		if k < 0 { //data line must have minimum 2 column separated ' ' space
			return -1.0
		}
		dept1, err = strconv.ParseFloat(s[:k], 64)
		if err != nil {
			return -1.0
		}
		j++
		if j == 2 {
			return math.Round((dept1-dept2)*10) / 10
		}
		dept2 = dept1
	}
	return 1.0
}

//setNull - change parameter NULL in WELL INFO section and in all logs
func (o *Las) setNull(aNull float64) error {
	for _, l := range o.Logs { //loop by logs
		for i := range l.log { //loop by dept step
			if l.log[i] == o.Null {
				l.log[i] = aNull
			}
		}
	}
	o.Null = aNull
	return nil
}

//TODO replace to function xStrUtil.ConvertStrCodePage
//вызывать после Scanner.Text()
func (o *Las) convertStrFromIn(s string) string {
	switch o.fromCodePage {
	case xlib.Cp866:
		s, _, _ = transform.String(charmap.CodePage866.NewDecoder(), s)
	case xlib.CpWindows1251:
		s, _, _ = transform.String(charmap.Windows1251.NewDecoder(), s)
	}
	return s
}

//TODO replace to function xStrUtil.ConvertStrCodePage
func (o *Las) convertStrToOut(s string) string {
	switch o.toCodePage {
	case xlib.Cp866:
		s, _, _ = transform.String(charmap.CodePage866.NewEncoder(), s)
	case xlib.CpWindows1251:
		s, _, _ = transform.String(charmap.Windows1251.NewEncoder(), s)
	}
	return s
}

//logByIndex - return log from map by Index
func (o *Las) logByIndex(i int) (*LasLog, error) {
	for _, v := range o.Logs {
		if v.Index == i {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("log with index: %v not present", i)
}

//NewLas - make new object Las class
//autodetect code page at load file
//code page to save by default is xlib.CpWindows1251
func NewLas(outputCP ...int) *Las {
	las := new(Las)
	las.Logs = make(map[string]LasLog)
	//las.warnings = make([]TWarning)
	if len(outputCP) > 0 {
		las.toCodePage = outputCP[0]
	} else {
		las.toCodePage = xlib.CpWindows1251
	}
	//mnemonic dictionary
	las.LogDic = nil
	//external log dictionary
	las.VocDic = nil
	//счётчик повторяющихся мнемоник, увеличивается каждый раз на 1, используется при переименовании мнемоники
	las.iDuplicate = 0
	return las
}

//analize first char after ~
//~V - section vertion
//~W - well info section
//~C - curve info section
//~A - data section
func (o *Las) selectSection(r rune) int {
	switch r {
	case 86: //V
		return lasSecVertion //version section
	case 118: //v
		return lasSecVertion //version section
	case 87: //W
		return lasSecWellInfo //well info section
	case 119: //w
		return lasSecWellInfo //well info section
	case 67: //C
		return lasSecCurInfo //curve section
	case 99: //c
		return lasSecCurInfo //curve section
	case 65: //A
		return lasSecData //data section
	case 97: //a
		return lasSecData //data section
	default:
		return lasSecIgnore
	}
}

//make test of loaded well info section
func (o *Las) testWellInfo() error {
	//STEP
	if o.Step == 0.0 {
		o.Step = o.getStepFromData(o.FileName)
		if o.Step < 0 {
			return errors.New("invalid STEP parameter, equal 0. and invalid step in data")
		}
		o.addWarning(TWarning{directOnRead, lasSecWellInfo, -1, fmt.Sprintf("invalid STEP parameter, equal 0. replace to %4.3f", o.Step)})
	}
	if o.Null == 0.0 {
		o.Null = Cfg.Null
		o.addWarning(TWarning{directOnRead, lasSecWellInfo, -1, fmt.Sprintf("invalid NULL parameter, equal 0. replace to %4.3f", o.Null)})
	}
	if math.Abs(o.Stop-o.Strt) < 0.2 {
		//TODO replace error to warning. invalid STRT and/or STOP replace to actual
		return errors.New("invalid STRT or STOP parameter, too small distance")
	}
	return nil
}

//Wraped - return true if las file have WRAP == YES
func (o *Las) Wraped() bool {
	return (strings.Index(o.Wrap, "Y") >= 0)
}

func (o *Las) addWarning(w TWarning) {
	if o.countWarning < Cfg.maxWarningCount {
		//w.desc = fmt.Sprintf("file: '%s' > "+w.desc, o.FileName)
		o.warnings = append(o.warnings, w)
		o.countWarning++
		if o.countWarning == Cfg.maxWarningCount {
			o.warnings = append(o.warnings, TWarning{0, 0, 0, "*maximum count* of warning reached, change parameter 'maxWarningCount' in 'glas.ini'"})
		}
	}
}

//return Mnemonic from dictionary by Log Name
//if Mnemonic not found return empty string ""
func (o *Las) getMnemonic(logName string) string {
	if (o.LogDic == nil) || (o.VocDic == nil) {
		return "-"
	}
	v, ok := (*o.LogDic)[logName]
	if ok { //GOOD - название каротажа равно мнемонике
		return logName
	}
	v, ok = (*o.VocDic)[logName]
	if ok { //POOR - название каротажа загружаемого файла найдено в словаре подстановок, мнемоника найдена
		return v
	}
	return ""
}

//Разбор одной строки с мнемоникой каротажа
//Разбираем в переменную l а потом сохраняем в map
//Каждый каротаж характеризуется тремя именами
//iName    - имя каротажа в исходном файле, может повторятся
//Name     - ключ в map хранилище, повторятся не может. если в исходном есть повторение, то Name строится добавлением к iName индекса
//Mnemonic - мнемоника, берётся из словаря. если в словаре не найдено, то оставляем iName
func (o *Las) readCurveParam(s string) error {
	l := LasLog{}
	err := l.fromString(s)
	if err != nil {
		return err
	}
	l.iName = l.Name
	l.Mnemonic = o.getMnemonic(l.iName)
	if _, ok := o.Logs[l.Name]; ok {
		o.iDuplicate++
		s = fmt.Sprintf("%v", o.iDuplicate)
		l.Name += s
	}
	l.Index = len(o.Logs)
	//m := o.getCountPoint()
	m := o.expPoints
	l.dept = make([]float64, m)
	l.log = make([]float64, m)
	o.Logs[l.Name] = l
	return nil
}

//оцениваем количество точек в файле
func (o *Las) getCountPoint() int {
	var m int
	if math.Abs(o.Stop) > math.Abs(o.Strt) {
		m = int((o.Stop-o.Strt)/o.Step) + 2
	} else {
		m = int((o.Strt-o.Stop)/o.Step) + 2
	}
	if m < 0 {
		m = -m
	}
	return m
}

//loadHeader - read las file and load all section before dAta ~A
/*  secName: 0 - empty, 1 - Version, 2 - Well info, 3 - Curve info, 4 - A data
1. читаем строку
2. если коммент или пустая в игнор
3. если начало секции, определяем какой
4. если началась секция данных заканчиваем
5. читаем одну строку (это один параметер из известной нам секции) */
func (o *Las) loadHeader(fileName string) error {
	o.FileName = fileName
	iFile, err := os.Open(fileName) // open file to READ
	if err != nil {
		return errors.New("file: " + os.Args[1] + " can't open to read")
	}
	defer iFile.Close()
	s := ""
	secNum := 0
	iScanner := bufio.NewScanner(iFile)
	for i := 0; iScanner.Scan(); i++ {
		s = strings.TrimSpace(iScanner.Text())
		if (len(s) == 0) || (s[0] == '#') {
			continue
		}
		s = o.convertStrFromIn(s)
		if s[0] == '~' { //start new section
			secNum = o.selectSection(rune(s[1]))
			if secNum == lasSecCurInfo { //enter to Curve section.
				//проверка корректности данных секции WELL INFO перез загрузкой кривых и данных
				err = o.testWellInfo() //STEP != 0, NULL != 0, STRT & STOP
				if err != nil {
					return err
				}
				//возможно значение STEP изменилось... оцениваем количество точек и сохраняем
				o.expPoints = o.getCountPoint()
			}
			if secNum == lasSecData {
				break // dAta section read after //exit from for
			}
		} else {
			err = o.readParameter(s, secNum) //if not comment, not empty and not new section => parameter, read it
			if err != nil {
				o.addWarning(TWarning{directOnRead, secNum, -1, fmt.Sprintf("while process parameter: '%s' occure error: %v", s, err)})
			}
		}
	}
	return nil
}

//Open - load las file
func (o *Las) Open(fileName string) (int, error) {
	var err error

	if !xlib.FileExists(fileName) {
		return 0, errors.New("Las.Open() error: file " + fileName + " not exist")
	}
	o.fromCodePage, err = xlib.CodePageDetect(fileName, "~A")
	if err != nil {
		return 0, err
	}

	//open and close file
	err = o.loadHeader(fileName)
	if err != nil {
		return 0, err
	}

	if strings.Index(o.Wrap, "Y") >= 0 {
		o.addWarning(TWarning{directOnRead, lasSecData, -1, "WRAP = YES, file ignored"})
		return 0, nil
	}

	pos, iScanner, err := xlib.SeekFileToString(fileName, "~A")
	if pos > 0 {
		//pos in file at line "~A..."
		o.currentLine = pos
		return o.readDataSec(iScanner)
	}
	return 0, err
}

func (o *Las) readParameter(s string, secNum int) error {
	switch secNum {
	case lasSecVertion:
		return o.readVersionParam(s)
	case lasSecWellInfo:
		return o.readWellParam(s)
	case lasSecCurInfo:
		return o.readCurveParam(s)
	}
	return nil
}

func (o *Las) readVersionParam(s string) error {
	var err error
	p := LasParam{}
	p.fromString(s)
	switch p.Name {
	case "VERS":
		o.Ver, err = strconv.ParseFloat(p.Val, 64)
	case "WRAP":
		o.Wrap = p.Val
	}
	return err
}

func (o *Las) readWellParam(s string) error {
	/*>>>>>>>>>>>>>>>>>>>>>>>> ver 1.0, 1.2
	  _____________STRT, STOP, STEP, NULL
	  STEP .M                             0.10 : Step
	  ^^^^  ^^^^                          ^^^^   ^^^
	  Name  Unit                          Val    Desc
	  _____________OTHER PARAMETER IN WELL INFO SECTION
	  WELL .                              WELL : 12-Karas
	  ^^^^  ^^^^                          ^^^^   ^^^
	  Name  Unit                         >Val<    Desc

	  >>>>>>>>>>>>>>>>>>>>>>>> ver 2.0
	  _____________STRT, STOP, STEP, NULL
	  STEP .M                             0.10 : Step
	  ^^^^  ^^^^                          ^^^^   ^^^
	  Name  Unit                          Val    Desc
	  _____________OTHER PARAMETER IN WELL INFO SECTION
	  WELL .                          12-Karas : WELL
	  ^^^^  ^^^^                          ^^^^   ^^^
	  Name  Unit                          Val    Desc*/
	var err error
	p := LasParam{}
	err = p.fromString(s)
	if err != nil {
		return err
	}
	switch p.Name {
	case "STRT":
		o.Strt, err = strconv.ParseFloat(p.Val, 64)
	case "STOP":
		o.Stop, err = strconv.ParseFloat(p.Val, 64)
	case "STEP":
		o.Step, err = strconv.ParseFloat(p.Val, 64)
	case "NULL":
		o.Null, err = strconv.ParseFloat(p.Val, 64)
	case "WELL":
		if o.Ver < 2.0 {
			o.Well = p.Desc
		} else {
			o.Well = p.Val
		}
	}
	if err != nil {
		o.addWarning(TWarning{directOnRead, lasSecWellInfo, -1, fmt.Sprintf("detected param: %v, unit:%v, value: %v\n", p.Name, p.Unit, p.Val)})
	}
	return err
}

//expandDept - if actually data points exceeds
func (o *Las) expandDept(d *LasLog) {
	//actual number of points more then expected
	o.addWarning(TWarning{directOnRead, lasSecData, o.currentLine, "actual number of data lines more than expected, check: STRT, STOP, STEP"})
	o.addWarning(TWarning{directOnRead, lasSecData, o.currentLine, "expand number of points"})
	//ожидаем удвоения данных
	o.expPoints *= 2
	//need expand all logs
	//fmt.Printf("old dept len: %d, cap: %d\n", len(d.dept), cap(d.dept))

	newDept := make([]float64, o.expPoints, o.expPoints)
	copy(newDept, d.dept)
	d.dept = newDept

	newLog := make([]float64, o.expPoints, o.expPoints)
	copy(newLog, d.dept)
	d.log = newLog
	o.Logs[d.Name] = *d

	//fmt.Printf("new dept len: %d, cap: %d\n", len(d.dept), cap(d.dept))
	//loop over other logs
	n := len(o.Logs)
	var l *LasLog
	for j := 1; j < n; j++ {
		l, _ = o.logByIndex(j)
		newDept := make([]float64, o.expPoints, o.expPoints)
		copy(newDept, l.dept)
		l.dept = newDept

		newLog := make([]float64, o.expPoints, o.expPoints)
		copy(newLog, l.log)
		l.log = newLog
		o.Logs[l.Name] = *l
	}
}

//~ASCII Log Data
// 1419.2000 -9999.000 -9999.000     2.186     2.187
// 1419.3000 -9999.000 1.1   2.203     2.205
func (o *Las) readDataSec(iScanner *bufio.Scanner) (int, error) {
	var (
		//m    int
		v    float64
		err  error
		d    *LasLog
		l    *LasLog
		dept float64
		i    int
	)
	o.currentLine++
	n := len(o.Logs)
	d, _ = o.logByIndex(0) //dept log
	s := ""
	for i = 0; iScanner.Scan(); i++ {
		o.currentLine++
		if i == o.expPoints { //o.expPoints - оценённое количество точек исходя из шага
			o.expandDept(d)
		}

		s = strings.TrimSpace(iScanner.Text())

		//first column is DEPT
		k := strings.IndexRune(s, ' ')
		if k < 0 { //line must have n+1 column and n separated spaces block (+1 becouse first column DEPT)
			o.addWarning(TWarning{directOnRead, lasSecData, o.currentLine, fmt.Sprintf("line: %d is empty, ignore", o.currentLine)})
			i--
			continue
		}
		dept, err = strconv.ParseFloat(s[:k], 64)
		if err != nil {
			o.addWarning(TWarning{directOnRead, lasSecData, o.currentLine, fmt.Sprintf("first column '%s' not numeric, ignore", s[:k])})
			i--
			continue
		}
		//fmt.Printf("dept len: %d, cap: %d; i = %d, m = %d\n", len(d.dept), cap(d.dept), i, o.expPoints)
		d.dept[i] = dept
		if i > 1 {
			if math.Pow(((dept-d.dept[i-1])-(d.dept[i-1]-d.dept[i-2])), 2) > 0.1 {
				o.addWarning(TWarning{directOnRead, lasSecData, o.currentLine, fmt.Sprintf("step %5.2f ≠ previously step %5.2f", (dept - d.dept[i-1]), (d.dept[i-1] - d.dept[i-2]))})
				dept = d.dept[i-1] + o.Step
			}
			if math.Pow(((dept-d.dept[i-1])-o.Step), 2) > 0.1 {
				o.addWarning(TWarning{directOnRead, lasSecData, o.currentLine, fmt.Sprintf("actual step %5.2f ≠ global STEP %5.2f", (dept - d.dept[i-1]), o.Step)})
			}
		}
		s = strings.TrimSpace(s[k+1:]) //cut first column
		for j := 1; j < (n - 1); j++ {
			iSpace := strings.IndexRune(s, ' ')
			switch iSpace {
			case -1: //не все колонки прочитаны, а пробелов уже нет... пробуем игнорировать сроку заполняя оставшиеся каротажи NULLами
				o.addWarning(TWarning{directOnRead, lasSecData, o.currentLine, "not all column readed, set log value to NULL"})
			case 0:
				v = o.Null
			case 1:
				v, err = strconv.ParseFloat(s[:1], 64)
			default:
				v, err = strconv.ParseFloat(s[:iSpace], 64) //strconv.ParseFloat(s[:iSpace-1], 64)
			}
			if err != nil {
				o.addWarning(TWarning{directOnRead, lasSecData, o.currentLine, fmt.Sprintf("can't convert string: '%s' to number, set to NULL", s[:iSpace-1])})
				v = o.Null
			}
			l, err = o.logByIndex(j)
			if err != nil {
				o.nPoints = i
				return i, errors.New("internal ERROR, func (o *Las) readDataSec()::o.logByIndex(j) return error")
			}
			l.dept[i] = dept
			l.log[i] = v
			s = strings.TrimSpace(s[iSpace+1:])
		}
		//остаток - последняя колонка
		v, err = strconv.ParseFloat(s, 64)
		if err != nil {
			o.addWarning(TWarning{directOnRead, lasSecData, o.currentLine, "not all column readed, set log value to NULL"})
			v = o.Null
		}
		l, err = o.logByIndex(n - 1)
		if err != nil {
			o.nPoints = i
			return i, errors.New("internal ERROR, func (o *Las) readDataSec()::o.logByIndex(j) return error on last column")
		}
		l.dept[i] = dept
		l.log[i] = v
	}
	//i - actually readed lines and add (.) to data array
	o.nPoints = i
	return i, nil
}

//Save - save to file
//rewrite if file exist
//if useMnemonic == true then on save using std mnemonic on ~Curve section
//TODO las have field filename of readed las file, after save filename must update or not? warning occure on write for what file?
func (o *Las) Save(fileName string, useMnemonic ...bool) error {
	n := len(o.Logs) //log count
	if n <= 0 {
		return errors.New("logs not exist")
	}

	var f *os.File
	var err error
	if !xlib.FileExists(fileName) {
		err = os.MkdirAll(filepath.Dir(fileName), os.ModePerm)
		if err != nil {
			return errors.New("path: '" + filepath.Dir(fileName) + "' can't create >>" + err.Error())
		}
		f, err = os.Create(fileName) //Open file to WRITE
		if err != nil {
			return errors.New("file: '" + fileName + "' can't open to write >>" + err.Error())
		}
		defer f.Close()
	}

	fmt.Fprintf(f, _LasFirstLine)
	fmt.Fprintf(f, _LasVersion, o.Ver)
	fmt.Fprintf(f, _LasWrap)
	fmt.Fprintf(f, _LasCodePage)
	fmt.Fprintf(f, _LasWellInfoSec)
	fmt.Fprintf(f, _LasStrt, o.Strt)
	fmt.Fprintf(f, _LasStop, o.Stop)
	fmt.Fprintf(f, _LasStep, o.Step)
	fmt.Fprintf(f, _LasNull, o.Null)
	fmt.Fprintf(f, _LasWell, o.convertStrToOut(o.Well))
	fmt.Fprintf(f, _LasCurvSec)
	fmt.Fprintf(f, _LasCurvDept)

	s := _LasDataSec + " DEPT  |" //готовим строчку с названиями каротажей глубина всегда присутствует
	var l *LasLog
	for i := 1; i < n; i++ { //Пишем названия каротажей
		l, _ := o.logByIndex(i)
		if len(useMnemonic) > 0 {
			if len(l.Mnemonic) > 0 {
				l.Name = l.Mnemonic
			}
		}
		fmt.Fprintf(f, _LasCurvLine, o.convertStrToOut(l.Name), o.convertStrToOut(l.Unit)) //запись мнемоник в секции ~Curve
		s += " " + fmt.Sprintf("%-8s|", l.Name)                                            //Собираем строчку с названиями каротажей
	}

	//write data
	s += "\n"
	fmt.Fprintf(f, o.convertStrToOut(s))
	dept, _ := o.logByIndex(0)
	for i := 0; i < o.nPoints; i++ { //loop by dept (.)
		fmt.Fprintf(f, "%-9.3f ", dept.dept[i])
		for j := 1; j < n; j++ { //loop by logs
			l, err = o.logByIndex(j)
			if err != nil {
				o.addWarning(TWarning{directOnWrite, lasSecData, i, "logByIndex() return error, log not found, panic"})
				return errors.New("logByIndex() return error, log not found, panic")
			}
			fmt.Fprintf(f, "%-9.3f ", l.log[i])
		}
		fmt.Fprintln(f)
	}
	return nil
}
