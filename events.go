package main

type Events interface {
   probe() bool
   enable([]uint16)
   sample() []uint32
}

