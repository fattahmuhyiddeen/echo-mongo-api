package config

// DbName represents database name
var DbName = "idp"

// Env is environment of project. can be local / staging / production
var Env = "local"

//IsProduction to check whether environment is production
func IsProduction() bool {
	return Env == "production"
}
