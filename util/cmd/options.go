package cmd

import (
	"github.com/adriaandejonge/xld/util/intf"
)

type (
	Executer func(args intf.Command) (result string, err error)
	Finder   func(command string) (option Option, ok bool)

	OptionList []Option

	Option struct {
		Do          Executer
		Name        string
		Description string
		Help        string
		Permission  string // TODO []string instead?
		MinArgs     int
	}
)

func (optionList *OptionList) Finder() Finder {
	index := make(map[string]Option)

	for _, el := range *optionList {
		index[el.Name] = el
	}

	return func(command string) (option Option, ok bool) {
		val, ok := index[command]
		return val, ok
	}
}

func (optionList *OptionList) List() (options []Option) {
	options = make([]Option, 0)
	for _, el := range *optionList {
		// if permission ok
		options = append(options, el)
	}
	return
}
