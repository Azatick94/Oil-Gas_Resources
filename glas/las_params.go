package main

import (
	"errors"
	"strings"
)

//LasParam - class to store parameter from any section
type LasParam struct {
	iName    string
	Name     string
	Mnemonic string
	Unit     string
	Val      string
	Desc     string
}

func (o *LasParam) fromStringWithUnit(s string, iPoint int) error {
	//s === 'STEP .M         0.10 : Step'
	o.Name = strings.TrimSpace(s[:iPoint])
	s = s[iPoint+1:] // string now store all char after point
	//s === 'M         0.10 : Step'
	iSpace := strings.Index(s, " ")
	if iSpace < 0 {
		return errors.New("line : '" + s + "' is bad for Parameter, ignore")
	}
	o.Unit = ""
	if iSpace > 0 {
		o.Unit = s[:iSpace]
	}
	//s === '        0.10 : Step'
	s = strings.TrimSpace(s[iSpace+1:])
	//s === '0.10 : Step'
	iColon := strings.Index(s, ":")
	if iColon < 0 {
		return errors.New("line : '" + s + "' is bad for Parameter, must have ':'")
	}
	o.Val = strings.TrimSpace(s[:iColon])
	o.Desc = s[iColon+1:]
	return nil
}

func (o *LasParam) fromStringWithoutUnit(s string, iPoint int) error {
	//s === 'SP         : self'
	o.Name = strings.TrimSpace(s[:iPoint])
	o.Unit = ""
	o.Val = ""
	o.Desc = strings.TrimSpace(s[iPoint:])
	return nil
}

func (o *LasParam) fromString(s string) error {
	iPoint := strings.Index(s, ".")
	if iPoint < 0 {
		iPoint = strings.Index(s, ":")
		if iPoint < 0 {
			return errors.New("line : '" + s + "' is bad for Parameter, ignore")
		}
		return o.fromStringWithoutUnit(s, iPoint)
	}
	return o.fromStringWithUnit(s, iPoint)
}
