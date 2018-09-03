package main

import (
	"context"
	"github.com/vmware/govmomi/find"
	"log"
	"vdemo/src/utils"
	"vdemo/src/vmwareagent"
)

func main() {
	ctx := context.Background()

	c := vmwareagent.NewAuthenticatedClient(ctx)

	finder := find.NewFinder(c, true)
	datacenter, err := finder.Datacenter(ctx, "BJ-IDC-Dogfood")
	utils.CheckError(err)
	finder.SetDatacenter(datacenter)
	mingweiVm, err := finder.VirtualMachine(context.Background(), "3.5.x_ha_autotest1")
	utils.CheckError(err)
	if mingweiVm == nil {
		log.Fatal("search vm error")
	}

	//VmInfo := vmwareagent.GetVMInfo(ctx, mingweiVm)
	//
	//log.Printf("vmInfo:")
	//utils.JsonPutLine(VmInfo)

	vnics, vnets, vdss := vmwareagent.GetVmwareVmNetworkInfo(ctx,c,mingweiVm)
	for _,v:=range vnics {
		utils.JsonPutLine(v)
	}
	for _,v:=range vnets {
		utils.JsonPutLine(v)
	}
	for _,v:=range vdss {
		utils.JsonPutLine(v)
	}
}
