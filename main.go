package main

import (
   "flag"
   "fmt"
   "os"
   "path"
   "strconv"
   "time"
   "strings"
)

type Present struct {
   sensor Sensor
   mnemonics []string
}

const (
   Readings = 60 * 60 * 24 * 30
)

var (
   advanced   = flag.Bool("advanced", false, "list all events")
   listenAddr = flag.String("listenAddr", "localhost:8080", "HTTP service listen address")
   debug      = flag.Bool("debug", false, "print debugging output")
   events     = flag.String("events", "pgfault,pgmajfault,numa_hit,numa_miss,numa_foreign,numa_local,numa_other", "comma-separated list of events")
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

   sensors := []Present{
      {sensor: &Kernel{}},
      {sensor: &Numaconnect2{}},
   }

   // remove any sensors where probe fails
   for i := len(sensors)-1; i >= 0; i-- {
      if !sensors[i].sensor.probe() {
         sensors = append(sensors[:i], sensors[i+1:]...)
      }
   }

   if *debug {
      fmt.Printf("detected %v\n", sensors)
   }

   if *list {
      fmt.Printf("events supported:\n")

      for _, sensor := range sensors {
         for _, val := range *sensor.sensor.supported() {
            if *advanced || !val.advanced {
               fmt.Printf("%30s   %s\n", val.mnemonic, val.desc)
            }
         }
      }

      os.Exit(0)
   }

   delay := time.Duration(interval) * time.Second
   elems := strings.Split(*events, ",")
   total := 0

   // build a list of enabled events, storing index
   for i, _ := range sensors {
      var enabled []uint16

      for _, elem := range elems {
         for j, val := range *sensors[i].sensor.supported() {
            if val.mnemonic == elem {
               enabled = append(enabled, uint16(j))
               sensors[i].mnemonics = append(sensors[i].mnemonics, elem)
               total++
            }
         }
      }

      sensors[i].sensor.enable(enabled)
   }

   if total == 0 {
      fmt.Println("no matching events")
      os.Exit(0)
   }

   line := 0

   for {
      time.Sleep(delay)

      // print column headings
      if line == 0 {
         for _, sensor := range sensors {
            for _, mnemonic := range sensor.mnemonics {
               fmt.Printf("%s ", mnemonic)
            }
         }
         fmt.Println("")
      }
      line = (line + 1) % 25

      for _, sensor := range sensors {
         samples := sensor.sensor.sample()
         if len(samples) != len(sensor.mnemonics) {
            panic("internal error")
         }

         for i, mnemonic := range sensor.mnemonics {
            fmt.Printf("%*d ", len(mnemonic), samples[i])
         }
      }
      fmt.Println("")

      if flag.NArg() == 0 {
         break
      }
   }
}

func main() {
   flag.Parse()

   if os.Getuid() != 0 {
      fmt.Println("please run with sudo/root")
      os.Exit(1)
   }

   exe := path.Base(os.Args[0])
   if exe == "vmxstat" {
      vmxstat()
   }

   // FIXME for numascope
}
