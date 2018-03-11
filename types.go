package main

import (
	"os"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
)

type FloatingIPUpdatEvent struct {
	FloatingIPAddress string `json:"floating_ip_address"`
	FixedIPAddress    string `json:"fixed_ip_address"`
}

type FloatingIPPayLoad struct {
	FloatingIP FloatingIPUpdatEvent `json:"floatingip"`
}

type Message struct {
	EventType string            `json:"event_type"`
	Payload   FloatingIPPayLoad `json:"payload"`
}

type Controller struct {
	OpenstackComputeClient *gophercloud.ServiceClient
}

func newController() (*Controller, error) {
	opts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return nil, err
	}

	provider, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		return nil, err
	}

	region := os.Getenv("OS_REGION_NAME")
	if region == "" {
		region = "RegionOne"
	}

	client, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{Region: region})
	if err != nil {
		return nil, err
	}

	return &Controller{OpenstackComputeClient: client}, nil
}
