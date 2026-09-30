package cstructs

// This file contains structures for the API's responses

// Struct for a GET /<game-id> request
type SResponseS struct {
	Error 	*string	`json:"error"`
	Script 	string	`json:"script"`
}

// Struct for a GET /search?txt=... request
type SResponseSearch struct {
	Error 		*string		`json:"error"`
	Results 	[]string	`json:"results"`
}