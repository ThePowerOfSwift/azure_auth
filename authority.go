package main

import (
	"fmt"
)

type Authority struct {
	Host   string
	Tenant string
}

func (a Authority) String() string {
	return fmt.Sprintf("https://%s/%s%s", a.Host, a.Tenant, TokenPathConst)
}
