package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"log"
	"net/url"
)

const CONFIG_URL = "https://192.168.70.255"
const CONFIG_USERNAME = "vsphere.local\\administrator"
const CONFIG_PASSWORD = "!QAZ2wsx"
const CONFIG_SSLFLAG = true

func jsonPutLine(obj interface{}) {
	b, err := json.Marshal(obj)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	fmt.Println(string(b))
}

func NewClient(ctx context.Context) (*govmomi.Client, error) {
	u, err := soap.ParseURL(CONFIG_URL)
	if err != nil {
		return nil, err
	}
	u.User = url.UserPassword(CONFIG_USERNAME, CONFIG_PASSWORD)
	return govmomi.NewClient(ctx, u, CONFIG_SSLFLAG)
}

func NewAuthenticatedClient(ctx context.Context) *vim25.Client {
	u, err := soap.ParseURL(CONFIG_URL)
	if err != nil {
		return nil
	}
	u.User = url.UserPassword(CONFIG_USERNAME, CONFIG_PASSWORD)
	soapClient := soap.NewClient(u, true)
	vimClient, err := vim25.NewClient(ctx, soapClient)
	if err != nil {
		log.Fatal(err)
	}

	req := types.Login{
		This: *vimClient.ServiceContent.SessionManager,
	}

	req.UserName = u.User.Username()
	if pw, ok := u.User.Password(); ok {
		req.Password = pw
	}

	_, err = methods.Login(ctx, vimClient, &req)
	if err != nil {
		log.Fatal(err)
	}

	return vimClient
}

func getVMInfo(ctx context.Context, vm *object.VirtualMachine) mo.VirtualMachine {
	var o mo.VirtualMachine

	err := vm.Properties(ctx, vm.Reference(), nil, &o)
	if err != nil {
		log.Fatal(err)
	}

	return o
}

func main() {
	ctx := context.Background()

	c := NewAuthenticatedClient(ctx)

	//rootFolder := object.NewRootFolder(c)
	//
	//refs, err := rootFolder.Children(ctx)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//var shenzhenDC *object.Datacenter
	//
	//for _, ref := range refs {
	//	dc := ref.(*object.Datacenter)
	//	dcName, err := dc.ObjectName(ctx)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	if dcName == "Shenzhen" {
	//		shenzhenDC = dc
	//	}
	//}

	//folders, err := shenzhenDC.Folders(ctx)
	//if err != nil {
	//	log.Fatal(err)
	//}

	//hostFolders := folders.HostFolder
	// dsFolders := folders.DatastoreFolder
	//vmsFolders := folders.VmFolder

	// log.Println(dsFolders.InventoryPath)

	// dsRefs, err := dsFolders.Children(ctx)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// for _, dsRef := range dsRefs {
	// 	ds := dsRef.(*object.Datastore)
	// 	dsName, err := ds.ObjectName(ctx)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	log.Println(dsName)
	// }

	//get host folder
	//hsFolder, err := hostFolders.Children(ctx)
	//if err != nil {
	//	log.Fatal(err)
	//}

	//clusterComputeResource := hsFolder[0].(*object.ClusterComputeResource)

	// ccrName, err := clusterComputeResource.ObjectName(ctx)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// log.Println(ccrName)

	// hss, err := clusterComputeResource.Hosts(ctx)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// var hs17 *object.HostSystem
	// for _, hs := range hss {
	// 	hsName, err := hs.ObjectName(ctx)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	if hsName == "192.168.59.17" {
	// 		hs17 = hs
	// 		log.Println(hs17.String)
	// 		break
	// 	}
	// }

	//get vm
	// vmrefs, err := vmsFolders.Children(ctx)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	var mingweiVm *object.VirtualMachine
	// for _, vmref := range vmrefs {
	// 	if reflect.TypeOf(vmref).String() == "*object.VirtualMachine" {
	// 		vm := vmref.(*object.VirtualMachine)
	// 		vmName, err := vm.ObjectName(ctx)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		if vmName == "mingwei-test" {
	// 			mingweiVm = vm
	// 		}
	// 		log.Println(vmName)
	// 	} else {
	// 		vm := vmref.(*object.Folder)
	// 		vmName, err := vm.ObjectName(ctx)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}

	// 		log.Println(vmName)
	// 	}
	// }

	//si := object.NewSearchIndex(c)

	// dsRef, err := si.FindChild(ctx, dsFolders, "RemoteISCSIVMFS6")

	// dataStore := dsRef.(*object.Datastore)
	// fileManager := dataStore.NewFileManager(shenzhenDC, false)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// dataBrower, err := dataStore.Browser(ctx)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	finder := find.NewFinder(c, true)
	datacenter, err :=  finder.Datacenter(ctx, "Shenzhen")
	finder.SetDatacenter(datacenter)
	vm, err := finder.VirtualMachine(context.Background(), "yanbo_test_1")
	//vm, err := si.FindChild(ctx, vmsFolders, "yanbo/yanbo_dev")
	if err != nil {
		log.Fatal(err)
	}

	if vm != nil {
		mingweiVm = vm
	} else {
		log.Fatal("search vm error")
	}

	//mingweiVm.CreateSnapshot(ctx, "mingwei-snap-2", "test create snap 2", false, false)
	//log.Println("mingwei uuid:", mingweiVm.UUID)

	//var o mo.VirtualMachine
	//
	//err = mingweiVm.Properties(ctx, mingweiVm.Reference(), nil, &o)
	//
	//for _, device := range o.Config.Hardware.Device {
	//	d := device.GetVirtualDevice()
	//	jsonPutLine(d)
	//}

	//vdiskmanager := object.NewVirtualDiskManager(c)
	//uuid, err := vdiskmanager.QueryVirtualDiskUuid(ctx,"[17.20] yanbo_1/yanbo_1.vmdk", shenzhenDC)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//log.Println(uuid)

	//log.Println("mingwei uuid", o.Config.Uuid)
	//log.Println("cbt enabled:", *o.Config.ChangeTrackingEnabled)

	var vmconfig types.VirtualMachineConfigSpec
	enabledCBT := true
	vmconfig.ChangeTrackingEnabled = &enabledCBT

	_, err = mingweiVm.Reconfigure(ctx, vmconfig)
	if err != nil {
		log.Fatal(err)
	}

	//resourcePool := o.ResourcePool
	//var rp mo.ResourcePool
	//property.DefaultCollector(c).RetrieveOne(ctx, o.ResourcePool.Reference(), nil, &rp)
	//jsonPutLine(rp)

	//var eb mo.EnvironmentBrowser
	//property.DefaultCollector(c).RetrieveOne(ctx, o.EnvironmentBrowser.Reference(), nil, &eb)
	//jsonPutLine(eb)

	//var ds mo.Datastore
	//for _, dsref := range o.Datastore {
	//	property.DefaultCollector(c).RetrieveOne(ctx, dsref.Reference(), nil, &ds)
	//	jsonPutLine(ds)
	//}

	//var n mo.Network
	//for _, nref := range o.Network {
	//	property.DefaultCollector(c).RetrieveOne(ctx, nref.Reference(), nil, &n)
	//	jsonPutLine(n)
	//}


	// dsRefs, err := dsFolders.Children(ctx)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// log.Println("datastore list...")
	// for _, dsRef := range dsRefs {
	// 	ds := dsRef.(*object.Datastore)
	// 	dsName, err := ds.ObjectName(ctx)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	log.Println(dsName)
	// }

	// var remoteDs *object.Datastore
	// ds, err := si.FindChild(ctx, dsFolders, "RemoteISCSIVMFS6")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// if ds != nil {
	// 	remoteDs = ds.(*object.Datastore)
	// } else {
	// 	log.Fatal("search datastore error")
	// }

	// log.Println(remoteDs.ObjectName(ctx))

	var tempVm mo.VirtualMachine
	err = mingweiVm.Properties(ctx, mingweiVm.Reference(), nil, &tempVm)
	if err != nil {
		log.Println("get vm snapshot array error:", err)
	}

	jsonPutLine(tempVm)

	// if tempVm.Snapshot == nil {
	// 	log.Println("vm snapshot array empty")
	// }

	// log.Println(len(tempVm.Snapshot.RootSnapshotList))

	//hs, err := mingweiVm.HostSystem(ctx)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//netsys, err := hs.ConfigManager().NetworkSystem(ctx)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//var mns mo.HostNetworkSystem
	//
	//property.DefaultCollector(c).RetrieveOne(ctx, netsys.Reference(), nil, &mns)
	//jsonPutLine(mns)

	rootSnapShotRef := tempVm.RootSnapshot[0].Reference()
	changeAreaReq := &types.QueryChangedDiskAreas{
		This: mingweiVm.Reference(),
		Snapshot    : &rootSnapShotRef,
		DeviceKey   : 2000,
		StartOffset : 0,
		ChangeId: "*",
	}
	res, err := methods.QueryChangedDiskAreas(ctx,c, changeAreaReq)
	if err != nil {
		log.Fatal(err)
	}
	jsonPutLine(res)
}
