package edge

type IotContainer struct {
	iot *Iot
}

func (c *IotContainer) Set(iot *Iot) {
	c.iot = iot
}

func (c *IotContainer) Clear() {
	c.iot = nil
}

func (c *IotContainer) Get() *Iot {
	return c.iot
}
