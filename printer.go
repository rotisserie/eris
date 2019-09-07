package eris

type Printer interface {
	Print(e error)
}

func (defaulFormat *defaultFormatter) Print(e error) {

}
