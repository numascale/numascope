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
   dev.enable([]uint16{13})

   for {
      time.Sleep(delay)
      samples := dev.sample()
      fmt.Printf("%v\n", samples)

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
