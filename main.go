package main

import (
   "flag"
   "fmt"
   "os"
   "path"
   "strconv"
   "time"
)

type Reading struct {
   timestamp uint64 // nanoseconds
   val       uint64
}

const (
   Readings = 60 * 60 * 24 * 30
)

var (
   listenAddr = flag.String("listenAddr", "localhost:8080", "HTTP service listen address")
   debug      = flag.Bool("debug", false, "print debugging output")
   interval   = 1
)

func numastat() {
   switch {
   case flag.NArg() == 1:
      interval, _ = strconv.Atoi(flag.Arg(0))
   case flag.NArg() > 1:
      fmt.Println("usage: numascope [interval]")
      os.Exit(1)
   }

   dev := &Vmstat{} // Numachip2{}
   supported := dev.probe()

   if len(*supported) == 0 {
      fmt.Println("Numachip2 not detected")
      os.Exit(0)
   }

   delay := time.Duration(interval) * time.Second
   enabled := []uint16{13}
   dev.enable(enabled)
   line := 0

   for {
      time.Sleep(delay)

      // print column headings
      if line == 0 {
         for _, val := range enabled {
            fmt.Printf("%s ", (*supported)[val].mnemonic)
         }
         fmt.Println("")
      }
      line = (line + 1) % 25

      samples := dev.sample()
      for i, _ := range samples {
         name := (*supported)[enabled[i]].mnemonic
         fmt.Printf("%*d ", len(name), samples[i])
      }
      fmt.Println("")

      if flag.NArg() == 0 {
         break
      }
   }
}

func main() {
   flag.Parse()

   exe := path.Base(os.Args[0])
   if exe == "numastat" {
      numastat()
   }

   // FIXME for numascope
}
