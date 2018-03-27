package main

import f "github.com/meixinyun/common/cmd/design-pattern/factory"

func main() {

	var c f.WorkerCreator
	c = new(f.ProgramCreator)
	p := c.Create()
	taskProject := "ProgramProject"
	p.DoWork(&taskProject)

	c = new(f.FarmerCreator)
	fa := c.Create()
	taskFarmLand := "farmland"
	fa.DoWork(&taskFarmLand)

}
