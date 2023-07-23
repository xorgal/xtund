package internal

type IService struct {
	XTUND    string
	IPTABLES string
}

var Service = IService{
	XTUND:    "xtund.service",
	IPTABLES: "xtun-iptables.service",
}
