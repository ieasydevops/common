package factory

import "github.com/go/src/fmt"

type Worker interface {
	DoWork(task *string)
}

type WorkerCreator interface {
	Create() Worker
}

type Programmer struct {
}

func (this *Programmer)DoWork(task *string) {
	fmt.Printf("I'm a Program Worker. My work is coding,current project is:%s\n", *task)
}

type ProgramCreator struct {
}

func (this *ProgramCreator)Create() Worker {
	return new(Programmer)
}

type Farmer struct {

}

type FarmerCreator struct {

}

func (this *Farmer)DoWork(task *string) {
	fmt.Printf("I'm a Farmer. My work is doing farm work,current project is:%s\n", *task)
}

func (this *FarmerCreator)Create() Worker {
	return new(Farmer)
}


















