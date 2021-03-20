package nginx

import "errors"

var (
	errNotValidIp = errors.New("this IP looks invalid")
	errNotPrivateIp = errors.New("provided ip is not private")
	errNotOfficeIp = errors.New("provided ip is not belong to any of the office networks")
	errIpDenied = errors.New("you can't use this ip. Its usage denied by administration")
	errIpParse = errors.New("can't parse this IP. contact bot admin")
	errNetworkIP = errors.New("seems that you entered network address")
	errConfigExists = errors.New("config for this domain already exists")
)