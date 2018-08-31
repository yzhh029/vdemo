package vmwareagent

import (
	"context"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"log"
	"net/url"
	"vdemo/src/utils"
)

const CONFIG_URL = "https://192.168.70.255"
const CONFIG_USERNAME = "vsphere.local\\administrator"
const CONFIG_PASSWORD = "!QAZ2wsx"
const CONFIG_SSLFLAG = true

type VMwareDisk struct {
	Uuid         string
	Name         string
	Path         string
	ZbsVolumeId  string
	Duplicated   bool
	OutOfProtect bool
}
type VMwareNet struct {
	Uuid          string
	Name          string
	Type          string
	VlanId        uint32
	VsdUuid       string
	Gateway       string
	Subnetmask    string
	Host          []*VMwareVdsHost
	OriginalHosts map[string]*VMwareVdsHost
}
type VMwareVdsHost struct {
	DataIp         string
	Host           string
	HostUuid       string
	ManagementIp   string
	NicsAssociated []string
}

// connect vsphere with an auth client
func NewAuthenticatedClient(ctx context.Context) *vim25.Client {
	u, err := soap.ParseURL(CONFIG_URL)
	if err != nil {
		return nil
	}
	u.User = url.UserPassword(CONFIG_USERNAME, CONFIG_PASSWORD)
	soapClient := soap.NewClient(u, CONFIG_SSLFLAG)
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

// enable vm open cbt
func SetCBTEnabled(ctx context.Context, vm *object.VirtualMachine) error {
	var vmconfig types.VirtualMachineConfigSpec
	enabledCBT := true
	vmconfig.ChangeTrackingEnabled = &enabledCBT
	_, err := vm.Reconfigure(ctx, vmconfig)
	return err
}

//get an vm info
func GetVMInfo(ctx context.Context, vm *object.VirtualMachine) mo.VirtualMachine {
	var o mo.VirtualMachine
	err := vm.Properties(ctx, vm.Reference(), nil, &o)
	utils.CheckError(err)
	return o
}

// get vm disk info from vminfo
func GetVmwareVmDiskInfo(vm mo.VirtualMachine) map[int32]*VMwareDisk {
	disks := make(map[int32]*VMwareDisk)
	for _, device := range vm.Config.Hardware.Device {
		key := device.GetVirtualDevice().Key
		if key >= 2000 && key < 3000 {
			diskDevice := device.GetVirtualDevice()
			var disk VMwareDisk
			disk.Name = diskDevice.DeviceInfo.GetDescription().Label
			disk.Duplicated = false
			disk.OutOfProtect = false
			switch backing := diskDevice.Backing.(type) {
			case *types.VirtualDiskFlatVer2BackingInfo:
				disk.Path = backing.FileName
				disk.Uuid = backing.Uuid
			case *types.VirtualDiskSparseVer2BackingInfo:
				disk.Path = backing.FileName
				disk.Uuid = backing.Uuid
			case *types.VirtualDiskRawDiskMappingVer1BackingInfo:
				disk.Path = backing.FileName
				disk.Uuid = backing.Uuid
			case *types.VirtualDiskRawDiskVer2BackingInfo:
				disk.Path = backing.DescriptorFileName
				disk.Uuid = backing.Uuid
			default:
			}
			disks[key] = &disk
		}
	}
	return disks
}

func GetVmwareVmNetworkInfo(ctx context.Context, c *vim25.Client, vm *object.VirtualMachine) map[int32]*VMwareNet {
	vmInfo := GetVMInfo(ctx, vm)
	vnets := make(map[int32]*VMwareNet)
	for _, device := range vmInfo.Config.Hardware.Device {
		key := device.GetVirtualDevice().Key
		if key >= 4000 && key < 5000 {
			var net VMwareNet
			netDevice := device.(*types.VirtualEthernetCard)
			net.Name = netDevice.DeviceInfo.GetDescription().Label
			//backing := netDevice.Backing.(*types.VirtualEthernetCardNetworkBackingInfo)
			net.VlanId = uint32(key)
			net.Uuid = string(key)
			hostnetwork := GetHostSystemNetWork(ctx, c, vm)
			vss := hostnetwork.NetworkInfo.Vswitch
			var vswitch types.HostVirtualSwitch
			for _, vs := range vss {
				pgs := vs.Portgroup
				for _, pg := range pgs {
					if pg == "key-vim.host.PortGroup-"+net.Name {
						vswitch = vs
						break
					}
				}
			}
			net.Gateway = hostnetwork.IpRouteConfig.(*types.HostIpRouteConfig).DefaultGateway
			net.VsdUuid = vswitch.Key
			for _, pg := range hostnetwork.NetworkInfo.Portgroup {
				if pg.Key == "key-vim.host.PortGroup-"+net.Name {
					net.VlanId = uint32(pg.Spec.VlanId)
					break
				}
			}
			vnets[key] = &net
		}
	}
	return vnets
}

func GetDiskChangeIdFromSnapshot(
	ctx context.Context,
	client *vim25.Client,
	snapshotRef *types.ManagedObjectReference,
	key int32,
) string {
	collector := property.DefaultCollector(client)
	var snapshot mo.VirtualMachineSnapshot
	err := collector.RetrieveOne(ctx, snapshotRef.Reference(), nil, &snapshot)
	utils.CheckError(err)
	devices := object.VirtualDeviceList(snapshot.Config.Hardware.Device)
	disk := devices.FindByKey(key)
	switch backing := disk.GetVirtualDevice().Backing.(type) {
	case *types.VirtualDiskFlatVer2BackingInfo:
		return backing.ChangeId
	case *types.VirtualDiskSparseVer2BackingInfo:
		return backing.ChangeId
	case *types.VirtualDiskRawDiskMappingVer1BackingInfo:
		return backing.ChangeId
	case *types.VirtualDiskRawDiskVer2BackingInfo:
		return backing.ChangeId
	default:
		return "*"
	}
}

func GetHostSystemNetWork(ctx context.Context, c *vim25.Client, vm *object.VirtualMachine) mo.HostNetworkSystem {
	hs, err := vm.HostSystem(ctx)
	if err != nil {
		log.Fatal(err)
	}
	netsys, err := hs.ConfigManager().NetworkSystem(ctx)
	if err != nil {
		log.Fatal(err)
	}
	var mns mo.HostNetworkSystem
	property.DefaultCollector(c).RetrieveOne(ctx, netsys.Reference(), nil, &mns)
	return mns
}
