package main

import (
	"context"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
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

	VmInfo := vmwareagent.GetVMInfo(ctx, mingweiVm)

	log.Printf("vmInfo:")
	utils.JsonPutLine(VmInfo)

	hs, err := mingweiVm.HostSystem(ctx)
	if err != nil {
		log.Fatal(err)
	}
	netsys, err := hs.ConfigManager().NetworkSystem(ctx)
	if err != nil {
		log.Fatal(err)
	}
	var mns mo.HostNetworkSystem
	property.DefaultCollector(c).RetrieveOne(ctx, netsys.Reference(), nil, &mns)

	utils.JsonPutLine(mns)

}
