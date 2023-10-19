package webserver

import "errors"

var (
	ErrNotValidIp   = errors.New("[webserver] this IP looks invalid")
	ErrNotPrivateIp = errors.New("[webserver] provided ip is not private")
	ErrNotOfficeIp  = errors.New("[webserver] provided ip is not belong to any of the office networks")
	ErrIpDenied     = errors.New("[webserver] you can't use this ip. Its usage denied by administration")
	ErrIpParse      = errors.New("[webserver] can't parse this IP. contact bot admin")
	ErrNetworkIP    = errors.New("[webserver] seems that you entered network address")
	ErrConfigExists = errors.New("[webserver] config for this domain already exists")
)
