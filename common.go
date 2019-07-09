package main

import (
   "os"
   "fmt"
   "path"
   "runtime"
)

type Event struct {
   index    int16 // -1 means unindexed
   mnemonic string
   desc     string
   enabled  bool
}

type Sensor interface {
   // checks if sensor is present
   Present() bool
   Name() string
   Sources() uint
   // scans through and activates enabled events, and if discrete
   Enable(discrete bool)
   Lock()
   Unlock()
   // returns slice of events
   Events() []Event
   // returns headings of enabled events, accounting for discrete or not
   Headings() []string
   // returns samples
   Sample() []int64
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
