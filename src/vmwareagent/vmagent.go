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
	"net"
	"net/url"
	"strings"
	"vdemo/src/utils"
)

const CONFIG_URL = "https://192.168.30.255"
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

type VMwareNic struct {
	MacAddress           string
	Name                 string
	VlanUuid             string
	Vlans                []string
}

type VMwareNet struct {
	InternalVlanID interface{}
	UUID           string
	Name           string
	Type           uint32
	VlanID         uint32
	VdsUUID        string
	Gateway        string
	Subnetmask     string
}

type VMwareSwitch struct {
	UUID      string
	Name      string
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

func GetVmwareVmNetworkInfo(ctx context.Context, c *vim25.Client, vm *object.VirtualMachine) (map[int32]*VMwareNic, map[int32]*VMwareNet, map[int32]*VMwareSwitch) {
	vmInfo := GetVMInfo(ctx, vm)
	hostnetwork := GetHostSystemNetWork(ctx, c, vm)

	vnics := make(map[int32]*VMwareNic)
	vnets := make(map[int32]*VMwareNet)
	vdss := make(map[int32]*VMwareSwitch)
	for _, device := range vmInfo.Config.Hardware.Device {
		key := device.GetVirtualDevice().Key
		if key >= 4000 && key < 5000 {
			var vnic VMwareNic
			var vnet VMwareNet
			var vds VMwareSwitch
			switch netDevice := device.(type) {
			case *types.VirtualE1000:
				vnic.MacAddress = netDevice.MacAddress
				vnic.Name = netDevice.DeviceInfo.GetDescription().Label
				backing := netDevice.Backing.(*types.VirtualEthernetCardNetworkBackingInfo)
				vnet.Name = backing.DeviceName
			case *types.VirtualE1000e:
				vnic.MacAddress = netDevice.MacAddress
				vnic.Name = netDevice.DeviceInfo.GetDescription().Label
				backing := netDevice.Backing.(*types.VirtualEthernetCardNetworkBackingInfo)
				vnet.Name = backing.DeviceName
			case *types.VirtualVmxnet2:
				vnic.MacAddress = netDevice.MacAddress
				vnic.Name = netDevice.DeviceInfo.GetDescription().Label
				backing := netDevice.Backing.(*types.VirtualEthernetCardNetworkBackingInfo)
				vnet.Name = backing.DeviceName
			case *types.VirtualVmxnet3:
				vnic.MacAddress = netDevice.MacAddress
				vnic.Name = netDevice.DeviceInfo.GetDescription().Label
				backing := netDevice.Backing.(*types.VirtualEthernetCardNetworkBackingInfo)
				vnet.Name = backing.DeviceName
			case *types.VirtualPCNet32:
				vnic.MacAddress = netDevice.MacAddress
				vnic.Name = netDevice.DeviceInfo.GetDescription().Label
				backing := netDevice.Backing.(*types.VirtualEthernetCardNetworkBackingInfo)
				vnet.Name = backing.DeviceName
			case *types.VirtualSriovEthernetCard:
				vnic.MacAddress = netDevice.MacAddress
				vnic.Name = netDevice.DeviceInfo.GetDescription().Label
				backing := netDevice.Backing.(*types.VirtualEthernetCardNetworkBackingInfo)
				vnet.Name = backing.DeviceName
			default:
			}

			var vswitch types.HostVirtualSwitch
			var vswitchName string
			portgroups := hostnetwork.NetworkInfo.Portgroup
			for _, portgroup := range portgroups {
				if portgroup.Key == "key-vim.host.PortGroup-"+vnet.Name {
					vnet.VlanID = uint32(portgroup.Spec.VlanId)
					vnet.UUID = portgroup.Key
					vnic.VlanUuid = portgroup.Key
					vswitchName = portgroup.Vswitch
					break
				}
			}
			for _, vs := range hostnetwork.NetworkInfo.Vswitch {
				if vs.Key == vswitchName {
					vswitch = vs
					vds.UUID = vs.Key
					vds.Name = vs.Name
					break
				}
			}
			vnet.Gateway = hostnetwork.IpRouteConfig.(*types.HostIpRouteConfig).DefaultGateway
			vnet.VdsUUID = vswitch.Key

			nets := vmInfo.Guest.Net
			for _, net := range nets {
				if net.Network == vnet.Name {
					if isIPv6(net.IpAddress[0]){
						vnet.Subnetmask = getIpv6SubnetMask(int(net.IpConfig.IpAddress[0].PrefixLength))
					} else {
						vnet.Subnetmask = getIpv4SubnetMask(int(net.IpConfig.IpAddress[0].PrefixLength))
					}
				}
			}

			vnics[key] = &vnic
			vnets[key] = &vnet
			vdss[key] = &vds
		}
	}
	return vnics, vnets, vdss
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

func CreateSnapshot(ctx context.Context, vm *object.VirtualMachine, snapshotName string, snapshotDesc string) error {
	result, err := vm.CreateSnapshot(ctx, snapshotName, snapshotDesc, false, false)
	utils.CheckError(err)
	return result.Wait(ctx)
}

func DeleteSnapshot(ctx context.Context, vm *object.VirtualMachine, snapshotName string) error {
	consolidDate := true
	result, err := vm.RemoveSnapshot(ctx, snapshotName, true, &consolidDate)
	utils.CheckError(err)
	return result.Wait(ctx)
}

func GetSnapshotByName(ctx context.Context, c *vim25.Client, vm *object.VirtualMachine, snapshotName string) mo.VirtualMachineSnapshot {
	snapshotRef, err := vm.FindSnapshot(ctx, snapshotName)
	utils.CheckError(err)
	var snapshot mo.VirtualMachineSnapshot
	err = property.DefaultCollector(c).RetrieveOne(ctx, snapshotRef.Reference(), nil, &snapshot)
	utils.CheckError(err)
	return snapshot
}

func GetSnapshotIncrementData(ctx context.Context, c *vim25.Client, vm *object.VirtualMachine, snapshotRef *types.ManagedObjectReference, changeId string, deviceKey int32) []types.DiskChangeExtent {
	changeAreaReq := &types.QueryChangedDiskAreas{
		This:        vm.Reference(),
		Snapshot:    snapshotRef,
		DeviceKey:   deviceKey,
		StartOffset: 0,
		ChangeId:    changeId,
	}
	res, err := methods.QueryChangedDiskAreas(ctx, c, changeAreaReq)
	utils.CheckError(err)
	return res.Returnval.ChangedArea
}

func isIPv6(ip string) bool {
	if strings.Contains(ip, ":") {
		return true
	} else {
		return false
	}
}

func getSubnetMask(ones int, ipv6 bool) string {
	if ipv6 {
		return getIpv6SubnetMask(ones)
	} else {
		return getIpv4SubnetMask(ones)
	}
}


func getIpv6SubnetMask(ones int) string {
	m := net.CIDRMask(ones, 128)
	str := m.String()
	var subnetmask string
	ipv6Split(str, &subnetmask)
	return subnetmask
}

func getIpv4SubnetMask(ones int) string {
	m := net.CIDRMask(ones, 32)
	subnetMask := net.IPv4(m[0], m[1], m[2], m[3])
	return subnetMask.String()
}

func ipv6Split(key string,temp *string){
	if len(key) <= 4 {
		*temp = *temp+key
	}
	for i:=0;i<len(key);i++{
		if (i+1)%4==0 && i != len(key)-1{
			*temp = *temp+key[:i+1]+":"
			key = key[i+1:]
			ipv6Split(key,temp)
			break
		}
	}
}