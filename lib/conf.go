package ptp

import (
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// Conf is a global configuration of p2p daemon
type Conf struct {
	IPTool  string `yaml:"iptool"`
	TAPTool string `yaml:"taptool"`
	INFFile string `yaml:"inf_file"`
	MTU     int    `yaml:"mtu"`
	PMTU    bool   `yaml:"pmtu"`
}

// Load will read specified configuration file
// and unmarshal it into a struct
func (c *Conf) Load(filepath string) error {
	c.SetDefaults()
	if filepath == "" {
		return nil
	}
	yamlFile, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return fmt.Errorf("config parse failed: %s", err.Error())
	}
	return nil
}

// SetDefaults will fill default values for the configuration
func (c *Conf) SetDefaults() {
	c.IPTool = DefaultIPTool
	c.TAPTool = DefaultTAPTool
	c.INFFile = DefaultINFFile
	c.MTU = DefaultMTU
	c.PMTU = DefaultPMTU
}

// GetIPTool will return network configuration tool location
func (c *Conf) GetIPTool(preset string) string {
	if preset != "" {
		return preset
	}
	return c.IPTool
}

// GetTAPTool returns path to TAP interface creation script
// on Windows
func (c *Conf) GetTAPTool(preset string) string {
	if preset != "" {
		return preset
	}
	return c.TAPTool
}

// GetINFFile returns path to TAP driver inf file on Windows
func (c *Conf) GetINFFile(preset string) string {
	if preset != "" {
		return preset
	}
	return c.INFFile
}

// GetMTU will return MTU value
func (c *Conf) GetMTU(preset int) int {
	if preset != 0 {
		return preset
	}
	return c.MTU
}

// GetPMTU will return pmtu mode
func (c *Conf) GetPMTU() bool {
	return c.PMTU
}
