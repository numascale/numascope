package main

type Events interface {
   probe() bool
   enable([]uint16)
   sample() []uint32
}

type Event struct {
   index int16 // -1 means unindexed
   advanced bool
   mnemonic string
   desc string
}

