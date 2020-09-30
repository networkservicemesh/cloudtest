// Copyright (c) 2019-2020 Cisco Systems, Inc and/or its affiliates.
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

package config

const (
	Layer3 NetworkType = "layer3"
	Layer2 NetworkType = "layer2-individual"
	Mixed  NetworkType = "hybrid"
)

type NetworkType string

type NetworkConfig struct {
	Type      NetworkType    `yaml:"type"`
	PortVLANs map[string]int `yaml:"port-vlans"` // Port -> VLAN mapping for Layer2, Mixed network type
}

type HardwareDeviceConfig struct {
	HostName        string         `yaml:"host-name"` // Host name with variable substitutions supported.
	OperatingSystem string         `yaml:"os"`        // Operating system
	Name            string         `yaml:"name"`      // Host name prefix, will create ENV variable IP_HostName
	Network         *NetworkConfig `yaml:"network"`
}

type FacilityDeviceConfig struct {
	Plan         string `yaml:"plan"` // Plan
	BillingCycle string `yaml:"billing-cycle"`

	HardwareDeviceConfig `yaml:",inline"`
}

type HardwarePacketConfig struct {
	HardwareDevices      []*HardwareDeviceConfig `yaml:"hardware-devices"`      // A set of device configuration required to be created before starting cluster.
	HardwareReservations []string                `yaml:"hardware-reservations"` // A set of hardware reservations
}

type FacilityPacketConfig struct {
	Devices           []*FacilityDeviceConfig `yaml:"devices"`            // A set of device configuration required to be created before starting cluster.
	Facilities        []string                `yaml:"facilities"`         // A set of facility filters
	PreferredFacility string                  `yaml:"preferred-facility"` // A preferred facility key
}

type PacketConfig struct {
	SSHKey string `yaml:"ssh-key"` // A location of ssh key

	HardwarePacketConfig `yaml:",inline"`
	FacilityPacketConfig `yaml:",inline"`
}
