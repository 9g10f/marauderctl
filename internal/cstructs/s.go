package cstructs

type SResponseS struct {
	Error 	*string	`json:"error"`
	Script 	string	`json:"script"`
}

type SResponseSearch struct {
	Error 		*string		`json:"error"`
	Results 	[]string	`json:"results"`
}