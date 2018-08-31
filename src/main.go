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
	datacenter, err := finder.Datacenter(ctx, "Shenzhen")
	utils.CheckError(err)
	finder.SetDatacenter(datacenter)
	mingweiVm, err := finder.VirtualMachine(context.Background(), "mingwei_1")
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

	//diskInfo := vmwareagent.GetVmwareVmDiskInfo(VmInfo)
	//for _, disk := range diskInfo {
	//	log.Println("------------------")
	//	log.Println("disk.uuid=", disk.Uuid)
	//	log.Println("disk.name=", disk.Name)
	//	log.Println("disk.path=", disk.Path)
	//}

	//rootSnapShotRef := VmInfo.RootSnapshot[0].Reference()
	//changeAreaReq := &types.QueryChangedDiskAreas{
	//	This: mingweiVm.Reference(),
	//	Snapshot    : &rootSnapShotRef,
	//	DeviceKey   : 2000,
	//	StartOffset : 0,
	//	ChangeId: "*",
	//}
	//res, err := methods.QueryChangedDiskAreas(ctx,c, changeAreaReq)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//log.Printf("changeArea:")
	//utils.JsonPutLine(res)
}
