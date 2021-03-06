package builder

import (
	"bytes"
	"log"
	"os/exec"
	"strings"

	"github.com/ercole-io/ercole-agent-virtualization/config"
	"github.com/ercole-io/ercole-agent-virtualization/marshal"
	"github.com/ercole-io/ercole-agent-virtualization/model"
)

// fetchClusters return VMWare clusters from the given hyperVisor
func fetchClusters(hv config.Hypervisor) []model.ClusterInfo {
	var out []byte

	switch hv.Type {
	case "vmware":
		out = pwshFetcher("vmware.ps1", "-s", "cluster", hv.Endpoint, hv.Username, hv.Password)

	case "ovm":
		out = fetcher("ovm", "cluster", hv.Endpoint, hv.Username, hv.Password, hv.OvmUserKey, hv.OvmControl)

	default:
		log.Println("Hypervisor not supported:", hv.Type, "(", hv, ")")
		return make([]model.ClusterInfo, 0)
	}

	fetchedClusters := marshal.Clusters(out)
	for i := range fetchedClusters {
		fetchedClusters[i].Type = hv.Type
	}

	return fetchedClusters
}

// fetchVirtualMachines return VMWare virtual machines infos from the given hyperVisor
func fetchVirtualMachines(hv config.Hypervisor) []model.VMInfo {
	var vms []model.VMInfo

	switch hv.Type {
	case "vmware":
		out := pwshFetcher("vmware.ps1", "-s", "vms", hv.Endpoint, hv.Username, hv.Password)
		vms = marshal.VmwareVMs(out)

	case "ovm":
		out := fetcher("ovm", "vms", hv.Endpoint, hv.Username, hv.Password, hv.OvmUserKey, hv.OvmControl)
		vms = marshal.OvmVMs(out)

	default:
		log.Println("Hypervisor not supported:", hv.Type, "(", hv, ")")
		return make([]model.VMInfo, 0)
	}

	log.Printf("Got %d vms from hypervisor: %s", len(vms), hv.Endpoint)

	return vms
}

func pwshFetcher(fetcherName string, args ...string) []byte {
	baseDir := config.GetBaseDir()

	args = append([]string{baseDir + "/fetch/" + fetcherName}, args...)
	log.Println("Pwshfetching /usr/bin/pwsh/" + " " + strings.Join(args, " "))
	out, err := exec.Command("/usr/bin/pwsh", args...).Output()
	if err != nil {
		log.Print(string(out))
		log.Fatal(err)
	}

	return out
}

func fetcher(fetcherName string, args ...string) []byte {
	var (
		cmd    *exec.Cmd
		err    error
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	baseDir := config.GetBaseDir()
	log.Println("Fetching " + baseDir + "/fetch/" + fetcherName + " " + strings.Join(args, " "))

	cmd = exec.Command(baseDir+"/fetch/"+fetcherName, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()

	if len(stderr.Bytes()) > 0 {
		log.Print(string(stderr.Bytes()))
	}

	if err != nil {
		log.Fatal(err)
	}

	return stdout.Bytes()
}
