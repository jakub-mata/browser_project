package main

type DefaultValues struct {
	fontSize       uint16
	textAlign      string
	display        string
	textDecoration string
}

var defaultValuesMap = map[string]DefaultValues{
	"p": DefaultValues{
		fontSize: 16,
		display:  "block",
	},
}

func getBase(name string) DefaultValues {
	return defaultValuesMap[name]
}
