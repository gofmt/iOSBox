package handlers

import (
	"github.com/gookit/gcli/v3"
)

var DonationCommand = &gcli.Command{
	Name: "donation",
	Desc: "捐助作者，保持更新能力",
	Func: func(c *gcli.Command, args []string) error {

		return nil
	},
}
