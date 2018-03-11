package main

type FloatingIPUpdatEvent struct {
	FloatingIPAddress string `json:"floating_ip_address"`
	FixedIPAddress    string `json:"fixed_ip_address"`
}

type FloatingIPPayLoad struct {
	Floatingip FloatingIPUpdatEvent `json:"floatingip"`
}

type Message struct {
	EventType string            `json:"event_type"`
	Payload   FloatingIPPayLoad `json:"payload"`
}
