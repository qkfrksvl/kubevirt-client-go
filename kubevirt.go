package main

import (
	"fmt"
	"github.com/spf13/pflag"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"kubevirt.io/client-go/kubecli"
	"log"
	"regexp"
	"strconv"
)

const vmNamespace = "vm"

type VM struct {
	Name    string
	Power   string
	CPU     string
	Memory  int
	Network string
	Disk    int
	GuestOS string
	IP      string
}

func setClient() *kubecli.KubevirtClient {
	clientConfig := kubecli.DefaultClientConfig(&pflag.FlagSet{})

	vc, err := kubecli.GetKubevirtClientFromClientConfig(clientConfig)
	if err != nil {
		log.Fatalf("cannot obtain KubeVirt client: %v\n", err)
	}

	return &vc
}

func listVM(vc *kubecli.KubevirtClient) []VM {

	var vms []VM

	v, err := (*vc).VirtualMachine(vmNamespace).List(&k8smetav1.ListOptions{})
	if err != nil {
		klog.Error(err)
	}
	for _, k := range v.Items {

		vms = append(vms, getVM(vc, k.Name))
	}

	return vms
}

func getVM(vc *kubecli.KubevirtClient, name string) VM {
	var v VM

	vm, err := (*vc).VirtualMachine(vmNamespace).Get(name, &k8smetav1.GetOptions{})
	if err != nil {
		klog.Error(err)
	}

	v.Name = vm.Name
	v.Power = fmt.Sprint(vm.Status.PrintableStatus)

	v.CPU = fmt.Sprint(vm.Spec.Template.Spec.Domain.Resources.Requests.Cpu().String())
	tempMem := fmt.Sprint(vm.Spec.Template.Spec.Domain.Resources.Requests.Memory().String())
	v.Memory = parseInt(tempMem)
	v.Network = fmt.Sprint(vm.Spec.Template.Spec.Networks[0].Name)

	vmi, err := (*vc).VirtualMachineInstance(vmNamespace).Get(name, &k8smetav1.GetOptions{})
	if err != nil {
		klog.Error(err)
	}

	for _, d := range vmi.Status.VolumeStatus {
		if d.PersistentVolumeClaimInfo != nil {

			v.Disk += parseInt(d.PersistentVolumeClaimInfo.Capacity.Storage().String())
		}
	}

	v.GuestOS = vmi.Status.GuestOSInfo.PrettyName
	v.IP = vmi.Status.Interfaces[0].IP

	return v
}

func parseInt(s string) int {
	re := regexp.MustCompile("[0-9]+")
	tS := re.FindAllString(s, 1)
	tI, _ := strconv.Atoi(tS[0])
	return tI
}
