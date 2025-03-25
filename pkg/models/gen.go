package models

//go:generate rm -f ./*.xo.go
//go:generate goschema generate --out=./ --sql=./schemas/*.sql --extension=xo
