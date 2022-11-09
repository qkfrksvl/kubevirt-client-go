package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/pflag"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"kubevirt.io/client-go/kubecli"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type VM struct {
	Name    string
	Power   string
	CPU     uint32
	Memory  int
	Network string
	Disk    int
	GuestOS string
	IP      string
}

type vcGetVM struct {
	vc *kubecli.KubevirtClient
}
type vcGetVMs struct {
	vc *kubecli.KubevirtClient
}

func main() {

	// kubecli.DefaultClientConfig() prepares config using kubeconfig.
	// typically, you need to set env variable, KUBECONFIG=<path-to-kubeconfig>/.kubeconfig
	clientConfig := kubecli.DefaultClientConfig(&pflag.FlagSet{})

	// retrive default namespace.
	namespace, _, err := clientConfig.Namespace()
	if err != nil {
		log.Fatalf("error in namespace : %v\n", err)
	}

	// get the kubevirt client, using which kubevirt resources can be managed.
	virtClient, err := kubecli.GetKubevirtClientFromClientConfig(clientConfig)
	if err != nil {
		log.Fatalf("cannot obtain KubeVirt client: %v\n", err)
	}

	// Fetch list of VMs & VMIs
	_, err = virtClient.VirtualMachine(namespace).List(&k8smetav1.ListOptions{})
	if err != nil {
		log.Fatalf("cannot obtain KubeVirt vm list: %v\n", err)
	}
	_, err = virtClient.VirtualMachineInstance(namespace).List(&k8smetav1.ListOptions{})
	if err != nil {
		log.Fatalf("cannot obtain KubeVirt vmi list: %v\n", err)
	}

	GetVM(&virtClient, "test02")

	httpGetVM := vcGetVM{vc: &virtClient}
	httpGetVMs := vcGetVMs{vc: &virtClient}
	http.Handle("/vm/", &httpGetVM)
	http.Handle("/vms", &httpGetVMs)
	http.ListenAndServe(":8000", nil)

}

func GetVM(vClient *kubecli.KubevirtClient, name string) (rv *VM) {
	v := &VM{}

	defer func() {
		if err := recover(); err != nil {
			rv = v

		}
	}()

	vm, err := (*vClient).VirtualMachine("default").Get(name, &k8smetav1.GetOptions{})
	if err != nil {
		klog.Error(err)
	}

	v.Name = vm.Name

	v.Power = fmt.Sprint(vm.Status.PrintableStatus)
	v.CPU = vm.Spec.Template.Spec.Domain.CPU.Cores
	tempMem := vm.Spec.Template.Spec.Domain.Resources.Requests.Memory().String()
	v.Memory = parseInt(tempMem)
	v.Network = vm.Spec.Template.Spec.Networks[0].Name
	vmi, err := (*vClient).VirtualMachineInstance("default").Get(name, &k8smetav1.GetOptions{})
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

func GetVMs(vClient *kubecli.KubevirtClient) []*VM {
	vl := []*VM{}
	vm, err := (*vClient).VirtualMachine("default").List(&k8smetav1.ListOptions{})
	if err != nil {
		klog.Error(err)
	}
	for _, v := range vm.Items {

		vl = append(vl, GetVM(vClient, v.Name))
	}

	return vl
}

func parseInt(s string) int {
	re := regexp.MustCompile("[0-9]+")
	tS := re.FindAllString(s, 1)
	tI, _ := strconv.Atoi(tS[0])
	return tI
}

func (c *vcGetVM) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u := strings.Split(r.URL.Path, "/")

	json.NewEncoder(w).Encode(GetVM(c.vc, u[2]))
}

func (c *vcGetVMs) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	json.NewEncoder(w).Encode(GetVMs(c.vc))

}
