// Copyright (c) 2019 Cisco Systems, Inc and/or its affiliates.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package packet

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/packethost/packngo"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/networkservicemesh/cloudtest/pkg/config"
	"github.com/networkservicemesh/cloudtest/pkg/execmanager"
	"github.com/networkservicemesh/cloudtest/pkg/k8s"
	"github.com/networkservicemesh/cloudtest/pkg/providers"
	"github.com/networkservicemesh/cloudtest/pkg/shell"
	"github.com/networkservicemesh/cloudtest/pkg/utils"
)

const (
	installScript   = "install" // #1
	setupScript     = "setup"   // #2
	startScript     = "start"   // #3
	configScript    = "config"  // #4
	prepareScript   = "prepare" // #5
	stopScript      = "stop"    // #6
	cleanupScript   = "cleanup" // #7
	packetProjectID = "PACKET_PROJECT_ID"
)

type packetProvider struct {
	root    string
	indexes map[string]int
	sync.Mutex
	clusters    []packetInstance
	installDone map[string]bool
}

type packetInstance struct {
	installScript            []string
	setupScript              []string
	startScript              []string
	prepareScript            []string
	stopScript               []string
	manager                  execmanager.ExecutionManager
	root                     string
	id                       string
	configScript             string
	factory                  k8s.ValidationFactory
	validator                k8s.KubernetesValidator
	configLocation           string
	shellInterface           shell.Manager
	projectID                string
	packetAuthKey            string
	keyID                    string
	config                   *config.ClusterProviderConfig
	provider                 *packetProvider
	client                   *packngo.Client
	project                  *packngo.Project
	devices                  map[string]*packngo.Device
	sshKey                   *packngo.SSHKey
	params                   providers.InstanceOptions
	started                  bool
	keyIds                   []string
	virtualNetworkList       []packngo.VirtualNetwork
	hardwareReservationsList []string
	facilitiesList           []string
}

func (pi *packetInstance) GetID() string {
	return pi.id
}

func (pi *packetInstance) CheckIsAlive() error {
	if pi.started {
		return pi.validator.Validate()
	}
	return errors.New("cluster is not running")
}

func (pi *packetInstance) IsRunning() bool {
	return pi.started
}

func (pi *packetInstance) GetClusterConfig() (string, error) {
	if pi.started {
		return pi.configLocation, nil
	}
	return "", errors.New("cluster is not started yet")
}

func (pi *packetInstance) Start(timeout time.Duration) (string, error) {
	logrus.Infof("Starting cluster %s-%s", pi.config.Name, pi.id)
	var err error
	fileName := ""
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Set seed
	rand.Seed(time.Now().UnixNano())

	utils.ClearFolder(pi.root, true)

	// Process and prepare environment variables
	if err = pi.shellInterface.ProcessEnvironment(
		pi.id, pi.config.Name, pi.root, pi.config.Env, nil); err != nil {
		logrus.Errorf("error during processing environment variables %v", err)
		return "", err
	}

	// Do prepare
	if !pi.params.NoInstall {
		if fileName, err = pi.doInstall(ctx); err != nil {
			return fileName, err
		}
	}

	// Run start script
	if fileName, err = pi.shellInterface.RunCmd(ctx, "setup", pi.setupScript, nil); err != nil {
		return fileName, err
	}

	keyFile := pi.config.Packet.SSHKey
	if !utils.FileExists(keyFile) {
		// Relative file
		keyFile = path.Join(pi.root, keyFile)
		if !utils.FileExists(keyFile) {
			err = errors.New("failed to locate generated key file, please specify init script to generate it")
			logrus.Errorf(err.Error())
			return "", err
		}
	}

	if pi.client, err = packngo.NewClient(); err != nil {
		logrus.Errorf("failed to create Packet REST interface")
		return "", err
	}

	if err = pi.updateProject(); err != nil {
		return "", err
	}

	// Check and add key if it is not yet added.

	if pi.keyIds, err = pi.createKey(keyFile); err != nil {
		return "", err
	}

	var virtualNetworks *packngo.VirtualNetworkListResponse
	virtualNetworks, _, err = pi.client.ProjectVirtualNetworks.List(pi.projectID, &packngo.ListOptions{})
	if err != nil {
		return "", err
	}
	pi.virtualNetworkList = virtualNetworks.VirtualNetworks

	if pi.hardwareReservationsList, err = pi.findHardwareReservations(); err != nil {
		return "", err
	}
	for _, devCfg := range pi.config.Packet.HardwareDevices {
		var device *packngo.Device
		device, err = pi.createHardwareDevice(devCfg)
		if err != nil {
			return "", nil
		}
		pi.devices[devCfg.Name] = device
	}

	if pi.facilitiesList, err = pi.findFacilities(); err != nil {
		return "", err
	}
	for _, devCfg := range pi.config.Packet.Devices {
		var device *packngo.Device
		device, err = pi.createFacilityDevice(devCfg)
		if err != nil {
			return "", nil
		}
		pi.devices[devCfg.Name] = device
	}

	// All devices are created so we need to wait for them to get alive.
	if err = pi.waitDevicesStartup(ctx); err != nil {
		return "", err
	}
	// We need to add arguments

	pi.addDeviceContextArguments()

	printableEnv := pi.shellInterface.PrintEnv(pi.shellInterface.GetProcessedEnv())
	pi.manager.AddLog(pi.id, "environment", printableEnv)

	// Run start script
	if fileName, err = pi.shellInterface.RunCmd(ctx, "start", pi.startScript, nil); err != nil {
		return fileName, err
	}

	if err = pi.updateKUBEConfig(ctx); err != nil {
		return "", err
	}

	if pi.validator, err = pi.factory.CreateValidator(pi.config, pi.configLocation); err != nil {
		msg := fmt.Sprintf("Failed to start validator %v", err)
		logrus.Errorf(msg)
		return "", err
	}
	// Run prepare script
	if fileName, err = pi.shellInterface.RunCmd(ctx, "prepare", pi.prepareScript, []string{"KUBECONFIG=" + pi.configLocation}); err != nil {
		return fileName, err
	}

	// Wait a bit to be sure clusters are up and running.
	st := time.Now()
	err = pi.validator.WaitValid(ctx)
	if err != nil {
		logrus.Errorf("Failed to wait for required number of nodes: %v", err)
		return fileName, err
	}
	logrus.Infof("Waiting for desired number of nodes complete %s-%s %v", pi.config.Name, pi.id, time.Since(st))

	pi.started = true
	logrus.Infof("Starting are up and running %s-%s", pi.config.Name, pi.id)
	return "", nil
}

func (pi *packetInstance) updateKUBEConfig(context context.Context) error {
	if pi.configLocation == "" {
		pi.configLocation = pi.shellInterface.GetConfigLocation()
	}
	if pi.configLocation == "" {
		output, err := utils.ExecRead(context, "", strings.Split(pi.configScript, " "))
		if err != nil {
			err = errors.Wrap(err, "failed to retrieve configuration location")
			logrus.Errorf(err.Error())
		}
		pi.configLocation = output[0]
	}
	return nil
}

func (pi *packetInstance) addDeviceContextArguments() {
	for key, dev := range pi.devices {
		for _, n := range dev.Network {
			pub := "pub"
			if !n.Public {
				pub = "private"
			}
			pi.shellInterface.AddExtraArgs(fmt.Sprintf("device.%v.%v.%v.%v", key, pub, "ip", n.AddressFamily), n.Address)
			pi.shellInterface.AddExtraArgs(fmt.Sprintf("device.%v.%v.%v.%v", key, pub, "gw", n.AddressFamily), n.Gateway)
			pi.shellInterface.AddExtraArgs(fmt.Sprintf("device.%v.%v.%v.%v", key, pub, "net", n.AddressFamily), n.Network)
		}
	}
}

func (pi *packetInstance) waitDevicesStartup(context context.Context) error {
	_, fileID, err := pi.manager.OpenFile(pi.id, "wait-nodes")
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(fileID)
	defer func() { _ = fileID.Close() }()
	for {
		alive := map[string]*packngo.Device{}
		for key, d := range pi.devices {
			var updatedDevice *packngo.Device
			updatedDevice, _, err := pi.client.Devices.Get(d.ID, &packngo.GetOptions{})
			if err != nil {
				logrus.Errorf("%v-%v Error accessing device Error: %v", pi.id, d.ID, err)
				continue
			} else if updatedDevice.State == "active" {
				alive[key] = updatedDevice
			}
			msg := fmt.Sprintf("Checking status %v %v %v", key, d.ID, updatedDevice.State)
			_, _ = writer.WriteString(msg)
			_ = writer.Flush()
			logrus.Infof("%v-Checking status %v", pi.id, updatedDevice.State)
		}
		if len(alive) == len(pi.devices) {
			pi.devices = alive
			break
		}
		select {
		case <-time.After(10 * time.Second):
			continue
		case <-context.Done():
			_, _ = writer.WriteString(fmt.Sprintf("Timeout"))
			return errors.Wrap(context.Err(), "timeout")
		}
	}
	_, _ = writer.WriteString(fmt.Sprintf("All devices online"))
	_ = writer.Flush()
	return nil
}

func (pi *packetInstance) createHardwareDevice(devCfg *config.HardwareDeviceConfig) (device *packngo.Device, err error) {
	var devReq *packngo.DeviceCreateRequest
	devReq, err = pi.createRequest(devCfg)
	if err != nil {
		return nil, err
	}

	var response *packngo.Response
	for _, hr := range pi.hardwareReservationsList {
		devReq.HardwareReservationID = hr
		device, response, err = pi.client.Devices.Create(devReq)
		msg := fmt.Sprintf("HostName=%v\n%v - %v", devReq.Hostname, response, err)
		logrus.Infof(fmt.Sprintf("%s-%v", pi.id, msg))
		pi.manager.AddLog(pi.id, fmt.Sprintf("create-device-%s", devCfg.Name), msg)
		if err == nil || err != nil &&
			!strings.Contains(err.Error(), "has no provisionable") &&
			!strings.Contains(err.Error(), "Oh snap, something went wrong") {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	if devCfg.Network != nil {
		return pi.setupDeviceNetwork(device, devCfg.Network)
	}
	return device, nil
}

func (pi *packetInstance) createFacilityDevice(devCfg *config.FacilityDeviceConfig) (device *packngo.Device, err error) {
	var devReq *packngo.DeviceCreateRequest
	devReq, err = pi.createRequest(&devCfg.HardwareDeviceConfig)
	if err != nil {
		return nil, err
	}
	devReq.Plan = devCfg.Plan
	devReq.BillingCycle = devCfg.BillingCycle

	var response *packngo.Response
	for i := range pi.hardwareReservationsList {
		devReq.Facility = []string{pi.hardwareReservationsList[i]}
		device, response, err = pi.client.Devices.Create(devReq)
		msg := fmt.Sprintf("HostName=%v\n%v - %v", devReq.Hostname, response, err)
		logrus.Infof(fmt.Sprintf("%s-%v", pi.id, msg))
		pi.manager.AddLog(pi.id, fmt.Sprintf("create-device-%s", devCfg.Name), msg)
		if err == nil || err != nil &&
			!strings.Contains(err.Error(), "has no provisionable") &&
			!strings.Contains(err.Error(), "Oh snap, something went wrong") {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	if devCfg.Network != nil {
		return pi.setupDeviceNetwork(device, devCfg.Network)
	}
	return device, nil
}

func (pi *packetInstance) createRequest(devCfg *config.HardwareDeviceConfig) (*packngo.DeviceCreateRequest, error) {
	finalEnv := pi.shellInterface.GetProcessedEnv()

	environment := map[string]string{}
	for _, k := range finalEnv {
		key, value, err := utils.ParseVariable(k)
		if err != nil {
			return nil, err
		}
		environment[key] = value
	}
	var hostName string
	var err error
	if hostName, err = utils.SubstituteVariable(devCfg.HostName, environment, pi.shellInterface.GetArguments()); err != nil {
		return nil, err
	}

	return &packngo.DeviceCreateRequest{
		Hostname:       hostName,
		OS:             devCfg.OperatingSystem,
		ProjectID:      pi.projectID,
		ProjectSSHKeys: pi.keyIds,
	}, err
}

func (pi *packetInstance) setupDeviceNetwork(device *packngo.Device, netCfg *config.NetworkConfig) (*packngo.Device, error) {
	piAddLog := func(format string, a ...interface{}) {
		pi.manager.AddLog(pi.id, "setup-device-network", fmt.Sprintf(format, a...))
	}

	var err error
	defer func() {
		if err != nil {
			piAddLog("error: %v", err)
		}
	}()

	device, err = pi.client.DevicePorts.DeviceToNetworkType(device.ID, string(netCfg.Type))
	piAddLog("device to network type: %v -> %v", device.Hostname, netCfg.Type)
	if err != nil {
		return nil, err
	}

	for portName, vlanTag := range netCfg.PortVLANs {
		var port *packngo.Port
		port, err = pi.client.DevicePorts.GetPortByName(device.ID, portName)
		if err != nil {
			return nil, err
		}

		var vlan *packngo.VirtualNetwork
		vlan, err = pi.findVlan(vlanTag)
		if err != nil {
			return nil, err
		}

		var response *packngo.Response
		_, response, err = pi.client.DevicePorts.Assign(&packngo.PortAssignRequest{
			PortID:           port.ID,
			VirtualNetworkID: vlan.ID,
		})
		piAddLog("port to vlan: %v -> %v\n%v", portName, vlanTag, response)
		if err != nil {
			return nil, err
		}
	}

	return device, nil
}

func (pi *packetInstance) findVlan(vlanTag int) (*packngo.VirtualNetwork, error) {
	for i := range pi.virtualNetworkList {
		if vlan := &pi.virtualNetworkList[i]; vlan.VXLAN == vlanTag {
			return vlan, nil
		}
	}
	return nil, errors.Errorf("vlan not found: %v", vlanTag)
}

func (pi *packetInstance) findHardwareReservations() ([]string, error) {
	hardwareReservations, response, err := pi.client.HardwareReservations.List(pi.projectID, &packngo.ListOptions{})

	out := strings.Builder{}
	_, _ = out.WriteString(fmt.Sprintf("%v\n%v\n", response.String(), err))

	if err != nil {
		pi.manager.AddLog(pi.id, "list-hardware-reservations", out.String())
		return nil, err
	}

	var hardwareReservationsList []string
	for i := range hardwareReservations {
		if !hardwareReservations[i].Provisionable {
			continue
		}

		for _, hrr := range pi.config.Packet.HardwareReservations {
			if hrr == hardwareReservations[i].ID {
				hardwareReservationsList = append(hardwareReservationsList, hrr)
			}
		}
	}

	return hardwareReservationsList, nil
}

func (pi *packetInstance) findFacilities() ([]string, error) {
	facilities, response, err := pi.client.Facilities.List(&packngo.ListOptions{})

	out := strings.Builder{}
	_, _ = out.WriteString(fmt.Sprintf("%v\n%v\n", response.String(), err))

	if err != nil {
		pi.manager.AddLog(pi.id, "list-facilities", out.String())
		return nil, err
	}

	var facilitiesList []string
	for _, f := range facilities {
		facilityReqs := map[string]string{}
		for _, ff := range f.Features {
			facilityReqs[ff] = ff
		}

		found := true
		for _, ff := range pi.config.Packet.Facilities {
			if _, ok := facilityReqs[ff]; !ok {
				found = false
				break
			}
		}
		if found {
			facilitiesList = append(facilitiesList, f.Code)
		}
	}

	// Randomize facilities.
	ind := -1

	if pi.config.Packet.PreferredFacility != "" {
		for i, f := range facilitiesList {
			if f == pi.config.Packet.PreferredFacility {
				ind = i
				break
			}
		}
	}

	if ind != -1 {
		facilitiesList[ind], facilitiesList[0] = facilitiesList[0], facilitiesList[ind]
	}

	msg := fmt.Sprintf("List of facilities: %v %v", facilities, response)
	//logrus.Infof(msg)
	_, _ = out.WriteString(msg)
	pi.manager.AddLog(pi.id, "list-facilities", out.String())

	return facilitiesList, nil
}

func (pi *packetInstance) Destroy(timeout time.Duration) error {
	logrus.Infof("Destroying cluster  %s", pi.id)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if pi.client != nil {
		if pi.sshKey != nil {
			response, err := pi.client.SSHKeys.Delete(pi.sshKey.ID)
			pi.manager.AddLog(pi.id, "delete-sshkey", fmt.Sprintf("%v\n%v\n%v", pi.sshKey, response, err))
		}

		_, logFile, err := pi.manager.OpenFile(pi.id, "destroy-cluster")
		defer func() { _ = logFile.Close() }()
		if err != nil {
			return err
		}
		_, _ = logFile.WriteString(fmt.Sprintf("Starting Delete of cluster %v", pi.id))
		iteration := 0
		for {
			alive := map[string]*packngo.Device{}
			for key, device := range pi.devices {
				var updatedDevice *packngo.Device
				updatedDevice, _, err := pi.client.Devices.Get(device.ID, &packngo.GetOptions{})
				if err != nil {
					if iteration == 0 {
						msg := fmt.Sprintf("%v-%v Error accessing device Error: %v", pi.id, device.ID, err)
						logrus.Error(msg)
						_, _ = logFile.WriteString(msg)
					} // else, if not first iteration and there is no device, just continue.
					continue
				}
				if updatedDevice.State != "provisioning" && updatedDevice.State != "queued" {
					response, err := pi.client.Devices.Delete(device.ID)
					if err != nil {
						_, _ = logFile.WriteString(fmt.Sprintf("delete-device-error-%s => %v\n%v ", key, response, err))
						logrus.Errorf("%v Failed to delete device %v", pi.id, device.ID)
					} else {
						_, _ = logFile.WriteString(fmt.Sprintf("delete-device-success-%s => %v\n%v ", key, response, err))
						logrus.Infof("%v Packet delete device send ok %v", pi.id, device.ID)
					}
				}
				// Put as alive or some different state
				alive[key] = updatedDevice

				msg := fmt.Sprintf("Device status %v %v %v", key, device.ID, updatedDevice.State)
				_, _ = logFile.WriteString(msg)
				logrus.Infof("%v-%v", pi.id, msg)
			}
			iteration++
			if len(alive) == 0 {
				break
			}
			select {
			case <-time.After(10 * time.Second):
				continue
			case <-ctx.Done():
				msg := fmt.Sprintf("Timeout for destroying cluster devices %v %v", pi.devices, ctx.Err())
				_, _ = logFile.WriteString(msg)
				return errors.Errorf("err: %v", msg)
			}
		}
		msg := fmt.Sprintf("Devices destroy complete %v", pi.devices)
		_, _ = logFile.WriteString(msg)
		logrus.Infof("Destroy Complete: %v", pi.id)
	}
	return nil
}

func (pi *packetInstance) GetRoot() string {
	return pi.root
}

func (pi *packetInstance) doInstall(context context.Context) (string, error) {
	pi.provider.Lock()
	defer pi.provider.Unlock()
	if pi.installScript != nil && !pi.provider.installDone[pi.config.Name] {
		pi.provider.installDone[pi.config.Name] = true
		return pi.shellInterface.RunCmd(context, "install", pi.installScript, nil)
	}
	return "", nil
}

func (pi *packetInstance) updateProject() error {
	ps, response, err := pi.client.Projects.List(nil)

	out := strings.Builder{}
	_, _ = out.WriteString(fmt.Sprintf("%v\n%v\n", response, err))

	if err != nil {
		logrus.Errorf("Failed to list Packet projects")
	}

	for i := 0; i < len(ps); i++ {
		p := &ps[i]
		_, _ = out.WriteString(fmt.Sprintf("Project: %v\n %v", p.Name, p))
		if p.ID == pi.projectID {
			pp := ps[i]
			pi.project = &pp
		}
	}

	pi.manager.AddLog(pi.id, "list-projects", out.String())

	if pi.project == nil {
		err := errors.Errorf("%s - specified project are not found on Packet %v", pi.id, pi.projectID)
		logrus.Errorf(err.Error())
		return err
	}
	return nil
}

func (pi *packetInstance) createKey(keyFile string) ([]string, error) {
	today := time.Now()
	genID := fmt.Sprintf("%d-%d-%d-%s", today.Year(), today.Month(), today.Day(), utils.NewRandomStr(10))
	pi.keyID = "dev-ci-cloud-" + genID

	out := strings.Builder{}
	keyFileContent, err := utils.ReadFile(keyFile)
	if err != nil {
		_, _ = out.WriteString(fmt.Sprintf("Failed to read key file %s", keyFile))
		pi.manager.AddLog(pi.id, "create-key", out.String())
		logrus.Errorf("Failed to read file %v %v", keyFile, err)
		return nil, err
	}

	_, _ = out.WriteString(fmt.Sprintf("Key file %s readed ok", keyFile))

	keyRequest := &packngo.SSHKeyCreateRequest{
		ProjectID: pi.project.ID,
		Label:     pi.keyID,
		Key:       strings.Join(keyFileContent, "\n"),
	}
	sshKey, response, err := pi.client.SSHKeys.Create(keyRequest)

	responseMsg := ""
	if response != nil {
		responseMsg = response.String()
	}
	createMsg := fmt.Sprintf("Create key %v %v %v", sshKey, responseMsg, err)
	_, _ = out.WriteString(createMsg)

	keyIds := []string{}
	if sshKey == nil {
		// try to find key.
		sshKey, keyIds = pi.findKeys(&out)
	} else {
		logrus.Infof("%s-Create key %v (%v)", pi.id, sshKey.ID, sshKey.Key)
		keyIds = append(keyIds, sshKey.ID)
	}
	pi.sshKey = sshKey
	pi.manager.AddLog(pi.id, "create-sshkey", fmt.Sprintf("%v\n%v\n%v\n %s", sshKey, response, err, out.String()))

	if sshKey == nil {
		_, _ = out.WriteString(fmt.Sprintf("Failed to create ssh key %v %v", sshKey, err))
		pi.manager.AddLog(pi.id, "create-key", out.String())
		logrus.Errorf("Failed to create ssh key %v", err)
		return nil, err
	}
	return keyIds, nil
}

func (pi *packetInstance) findKeys(out io.StringWriter) (*packngo.SSHKey, []string) {
	sshKeys, response, err := pi.client.SSHKeys.List()
	if err != nil {
		_, _ = out.WriteString(fmt.Sprintf("List keys error %v %v\n", response, err))
	}
	var keyIds []string
	var sshKey *packngo.SSHKey
	for k := 0; k < len(sshKeys); k++ {
		kk := &sshKeys[k]
		if kk.Label == pi.keyID {
			sshKey = &packngo.SSHKey{
				ID:          kk.ID,
				Label:       kk.Label,
				URL:         kk.URL,
				User:        kk.User,
				Key:         kk.Key,
				FingerPrint: kk.FingerPrint,
				Created:     kk.Created,
				Updated:     kk.Updated,
			}
		}
		_, _ = out.WriteString(fmt.Sprintf("Added key key %v\n", kk))
		keyIds = append(keyIds, kk.ID)
	}
	return sshKey, keyIds
}

func (p *packetProvider) getProviderID(provider string) string {
	val, ok := p.indexes[provider]
	if ok {
		val++
	} else {
		val = 1
	}
	p.indexes[provider] = val
	return fmt.Sprintf("%d", val)
}

func (p *packetProvider) CreateCluster(config *config.ClusterProviderConfig, factory k8s.ValidationFactory,
	manager execmanager.ExecutionManager,
	instanceOptions providers.InstanceOptions) (providers.ClusterInstance, error) {
	err := p.ValidateConfig(config)
	if err != nil {
		return nil, err
	}
	p.Lock()
	defer p.Unlock()
	id := fmt.Sprintf("%s-%s", config.Name, p.getProviderID(config.Name))

	root := path.Join(p.root, id)

	clusterInstance := &packetInstance{
		manager:        manager,
		provider:       p,
		root:           root,
		id:             id,
		config:         config,
		configScript:   config.Scripts[configScript],
		installScript:  utils.ParseScript(config.Scripts[installScript]),
		setupScript:    utils.ParseScript(config.Scripts[setupScript]),
		startScript:    utils.ParseScript(config.Scripts[startScript]),
		prepareScript:  utils.ParseScript(config.Scripts[prepareScript]),
		stopScript:     utils.ParseScript(config.Scripts[stopScript]),
		factory:        factory,
		shellInterface: shell.NewManager(manager, id, config, instanceOptions),
		params:         instanceOptions,
		projectID:      os.Getenv(packetProjectID),
		packetAuthKey:  os.Getenv("PACKET_AUTH_TOKEN"),
		devices:        map[string]*packngo.Device{},
	}

	return clusterInstance, nil
}

// CleanupClusters - Cleaning up leaked clusters
func (p *packetProvider) CleanupClusters(ctx context.Context, config *config.ClusterProviderConfig,
	manager execmanager.ExecutionManager, instanceOptions providers.InstanceOptions) {
	if _, ok := config.Scripts[cleanupScript]; !ok {
		// Skip
		return
	}

	logrus.Infof("Starting cleaning up clusters for %s", config.Name)
	shellInterface := shell.NewManager(manager, fmt.Sprintf("%s-cleanup", config.Name), config, instanceOptions)

	p.Lock()
	// Do prepare
	if skipInstall := instanceOptions.NoInstall || p.installDone[config.Name]; !skipInstall {
		if iScript, ok := config.Scripts[installScript]; ok {
			_, err := shellInterface.RunCmd(ctx, "install", utils.ParseScript(iScript), config.Env)
			if err != nil {
				logrus.Warnf("Install command for cluster %s finished with error: %v", config.Name, err)
			} else {
				p.installDone[config.Name] = true
			}
		}
	}
	p.Unlock()

	_, err := shellInterface.RunCmd(ctx, "cleanup", utils.ParseScript(config.Scripts[cleanupScript]), config.Env)
	if err != nil {
		logrus.Warnf("Cleanup command for cluster %s finished with error: %v", config.Name, err)
	}
}

// NewPacketClusterProvider - create new packet provider.
func NewPacketClusterProvider(root string) providers.ClusterProvider {
	utils.ClearFolder(root, true)
	return &packetProvider{
		root:        root,
		clusters:    []packetInstance{},
		indexes:     map[string]int{},
		installDone: map[string]bool{},
	}
}

func (p *packetProvider) ValidateConfig(config *config.ClusterProviderConfig) error {
	if config.Packet == nil {
		return errors.New("packet configuration element should be specified")
	}

	isHardware := len(config.Packet.HardwareDevices) > 0
	isFacility := len(config.Packet.Devices) > 0
	if !isHardware && !isFacility {
		return errors.New("packet configuration devices should be specified")
	}

	if isHardware && len(config.Packet.HardwareReservations) == 0 {
		return errors.New("packet hardware reservations should be specified")
	}

	if isFacility && len(config.Packet.Facilities) == 0 {
		return errors.New("packet configuration facilities should be specified")
	}

	if _, ok := config.Scripts[configScript]; !ok {
		hasKubeConfig := false
		for _, e := range config.Env {
			if strings.HasPrefix(e, "KUBECONFIG=") {
				hasKubeConfig = true
				break
			}
		}
		if !hasKubeConfig {
			return errors.New("invalid config location")
		}
	}
	if _, ok := config.Scripts[startScript]; !ok {
		return errors.New("invalid start script")
	}

	for _, envVar := range config.EnvCheck {
		envValue := os.Getenv(envVar)
		if envValue == "" {
			return errors.Errorf("environment variable are not specified %s Required variables: %v", envValue, config.EnvCheck)
		}
	}

	envValue := os.Getenv("PACKET_AUTH_TOKEN")
	if envValue == "" {
		return errors.New("environment variable are not specified PACKET_AUTH_TOKEN")
	}

	envValue = os.Getenv("PACKET_PROJECT_ID")
	if envValue == "" {
		return errors.New("environment variable are not specified PACKET_AUTH_TOKEN")
	}

	return nil
}
