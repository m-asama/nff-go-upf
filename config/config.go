package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Global   *Global    `yaml:"global"`
	Sessions []*Session `yaml:"sessions"`
}

type Global struct {
	CpuList *string `yaml:"cpuList"`
	Local   *Local  `yaml:"local"`
	N6      *N6     `yaml:"n6"`
	N3N9    *N3N9   `yaml:"n3n9"`
}

type Local struct {
	Port      *int    `yaml:"port"`
	Address   *string `yaml:"address"`
	TeAddress *string `yaml:"teAddress"`
}

type N6 struct {
	VlanId  *int    `yaml:"vlanId"`
	Address *string `yaml:"address"`
}

type N3N9 struct {
	VlanId  *int    `yaml:"vlanId"`
	Address *string `yaml:"address"`
}

type Session struct {
	Fseid *Fseid `yaml:"fseid"`
	Pdrs  []*Pdr `yaml:"pdrs"`
	Fars  []*Far `yaml:"fars"`
	Qers  []*Qer `yaml:"qers"`
}

type Fseid struct {
	Seid    *int    `yaml:"seid"`
	Address *string `yaml:"address"`
}

type Pdr struct {
	Pdrid              *int  `yaml:"pdrid"`
	Precedence         *int  `yaml:"precedence"`
	Pdi                *Pdi  `yaml:"pdi"`
	OuterHeaderRemoval *bool `yaml:"outerHeaderRemoval"`
	Farid              *int  `yaml:"farid"`
	Qerids             []int `yaml:"qerids"`
}

type Pdi struct {
	SourceInterface *string    `yaml:"sourceInterface"`
	Fteid           *Fteid     `yaml:"fteid"`
	NetworkInstance *string    `yaml:"networkInstance"`
	UeIpAddress     *string    `yaml:"ueIpAddress"`
	SdfFilter       *SdfFilter `yaml:"sdfFilter"`
}

type SdfFilter struct {
	SourcePrefix      *string `yaml:"sourcePrefix"`
	DestinationPrefix *string `yaml:"destinationPrefix"`
	Protocol          *string `yaml:"protocol"`
	SourcePorts       *string `yaml:"sourcePorts"`
	DestinationPorts  *string `yaml:"destinationPorts"`
}

type Fteid struct {
	Teid    *int    `yaml:"teid"`
	Address *string `yaml:"address"`
}

type Far struct {
	Farid                *int                  `yaml:"farid"`
	ApplyAction          *int                  `yaml:"applyAction"`
	ForwardingParameters *ForwardingParameters `yaml:"forwardingParameters"`
}

type ForwardingParameters struct {
	DestinationInterface *string              `yaml:"destinationInterface"`
	NetworkInstance      *string              `yaml:"networkInstance"`
	OuterHeaderCreation  *OuterHeaderCreation `yaml:"outerHeaderCreation"`
}

type OuterHeaderCreation struct {
	Teid    *int    `yaml:"teid"`
	Address *string `yaml:"address"`
}

type Qer struct {
	Qerid      *int    `yaml:"qerid"`
	GateStatus *string `yaml:"gateStatus"`
	Mbr        *Mbr    `yaml:"mbr"`
	Qfi        *int    `yaml:"qfi"`
}

type Mbr struct {
	Ul *uint64 `yaml:"ul"`
	Dl *uint64 `yaml:"dl"`
}

func Parse(f string) (*Config, error) {
	content, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	Config := Config{}
	err = yaml.Unmarshal(content, &Config)
	if err != nil {
		return nil, err
	}
	return &Config, nil
}

func (c *Config) Debug() {
	if c == nil {
		fmt.Println("Config is nil")
		return
	}
	if c.Global == nil {
		fmt.Println("Config.Global is nil")
	} else {
		if c.Global.CpuList == nil {
			fmt.Println("Config.Global.CpuList is nil")
		} else {
			fmt.Println("Config.Global.CpuList:", *c.Global.CpuList)
		}
	}
	if c.Sessions == nil || len(c.Sessions) == 0 {
		fmt.Println("Config.Sessions is nil")
	} else {
		for i, s := range c.Sessions {
			if s.Fseid == nil {
				fmt.Println("Sessions[", i, "].Fseid is nil")
			} else {
				fmt.Println("Sessions[", i, "].Fseid:", *s.Fseid)
			}
		}
	}
}
