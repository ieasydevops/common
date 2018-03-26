package singleton

import (
	"sync"
	"fmt"
)

var (
	_singleton *Singleton
	once sync.Once
)

type Singleton interface{}

func GetInstance() *Singleton {
	once.Do(func() {
		_singleton = &Singleton{}
	})
	return _singleton
}

func (p Singleton) Singleton() {
	fmt.Print("....Singleton Create ")
}






