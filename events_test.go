package main

import (
   "fmt"
   "testing"
)

func TestMain(m *testing.M) {
   fmt.Println("TestMain")
   dev := &Numaconnect2{}

   if dev.probe() > 0 {
      dev.enable([]uint16{0x68, 0x80}, false)
      for i := 0; i < 3; i++ {
         _ = dev.sample()
      }
   } else {
      fmt.Println("Numachip2 not detected")
   }
}
