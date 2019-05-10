package main

import (
   "fmt"
   "testing"
)

func TestMain(m *testing.M) {
   fmt.Println("TestMain")
   dev := &Numaconnect2{}

   if dev.Present() {
      events := dev.Events()
      events[1].enabled = true
      events[3].enabled = true

      dev.Enable(true)

      for i := 0; i < 3; i++ {
         _ = dev.Sample()
      }
   } else {
      fmt.Println("Numachip2 not detected")
   }
}
