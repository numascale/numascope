package main

import (
   "os"
   "fmt"
   "path"
   "runtime"
)

type Event struct {
   index int16 // -1 means unindexed
//   advanced bool
   mnemonic string
   desc string
}

type Sensor interface {
   probe() uint
   name() string
   supported() *[]Event
   enable([]uint16, bool)
   sample() []uint64
}

type Reading struct {
   timestamp uint64 // nanoseconds
   val       uint64
}

// Checks if an error occurred
func validate(err error) {
   if err != nil {
      _, file, line, _ := runtime.Caller(1)
      _, leaf := path.Split(file)
      fmt.Printf("Failed with '%v' at %v:%v\n", err, leaf, line)
      os.Exit(1)
   }
}
