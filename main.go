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
   advanced   = flag.Bool("advanced", false, "list all events")
   listenAddr = flag.String("listenAddr", "localhost:8080", "HTTP service listen address")
   debug      = flag.Bool("debug", false, "print debugging output")
   events     = flag.String("events", "", "comma-separated list of events")
   list       = flag.Bool("list", false, "list detected events")
   interval   = 1
)

func vmxstat() {
   switch {
   case flag.NArg() == 1:
      interval, _ = strconv.Atoi(flag.Arg(0))
   case flag.NArg() > 1:
      fmt.Println("usage: vmxstat [interval]")
      os.Exit(1)
   }

   dev := &Vmstat{} // Numachip2{}
   supported := dev.probe()

   if *list {
      fmt.Printf("events supported:\n")

      for _, val := range *supported {
         if *advanced || !val.advanced {
            fmt.Printf("%30s   %s\n", val.mnemonic, val.desc)
         }
      }

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
   if exe == "vmxstat" {
      vmxstat()
   }

   // FIXME for numascope
}
