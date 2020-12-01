package main

const (
	defWarningCount = 10
	warningUNDEF    = 0
	directOnRead    = 1
	directOnWrite   = 2
)

//TWarning - class to store warning
type TWarning struct {
	direct  int    // 0 - undefine (warning_UNDEF), 1 - on read (direct_ON_READ), 2 - on write (direct_ON_WRITE)
	section int    // 0 - undefine (warning_UNDEF), lasSecVertion, lasSecWellInfo, lasSecCurInfo, lasSecData
	line    int    // number of line in source file
	desc    string // description of warning
}

//TLasWarnings - class to store and manipulate warnings
type TLasWarnings struct {
}
