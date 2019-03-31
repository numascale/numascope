package main

import (
   "fmt"
   "testing"
)

func TestMain(m *testing.M) {
   fmt.Println("TestMain")
   dev := &Numachip2{}

   if dev.probe() {
      dev.enable([]uint16{0x68, 0x80})
      for i := 0; i < 3; i++ {
         _ = dev.sample()
      }
   } else {
      fmt.Println("Numachip2 not detected")
   }
}
