package nginx

import "errors"

var (
	errNotValidIp   = errors.New("[nginx] this IP looks invalid")
	errNotPrivateIp = errors.New("[nginx] provided ip is not private")
	errNotOfficeIp  = errors.New("[nginx] provided ip is not belong to any of the office networks")
	errIpDenied     = errors.New("[nginx] you can't use this ip. Its usage denied by administration")
	errIpParse      = errors.New("[nginx] can't parse this IP. contact bot admin")
	errNetworkIP    = errors.New("[nginx] seems that you entered network address")
	errConfigExists = errors.New("[nginx] config for this domain already exists")
)
