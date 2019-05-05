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
   units int
   enabled []*Event
}

const (
   Readings = 60 * 60 * 24 * 30
)

var (
// TODO enable advanced when there is useful discrimitation
//   advanced   = flag.Bool("advanced", false, "list all events")
   listenAddr = flag.String("listenAddr", "localhost:80", "web service listen address and port")
   debug      = flag.Bool("debug", false, "print debugging output")
   events     = flag.String("events", "pgfault,pgmajfault,numa_hit,numa_miss,numa_foreign,numa_local,numa_other", "comma-separated list of events")
   list       = flag.Bool("list", false, "list events available on this host")
   discrete   = flag.Bool("discrete", false, "report events per unit, rather than average")
   interval   = 1
   sensors    = []Present{
      {sensor: &Kernel{}},
      {sensor: &Numaconnect2{}},
   }
)

func usage() {
   fmt.Println("usage: vmxstat [interval]")
   os.Exit(1)
}

func vmxstat() {
   switch {
   case flag.NArg() == 1:
      var err error
      interval, err = strconv.Atoi(flag.Arg(0))
      if err != nil {
         usage()
      }
   case flag.NArg() > 1:
      usage()
   }

   if *debug {
      fmt.Printf("detected %v\n", sensors)
   }

   if *list {
      for _, sensor := range sensors {
         fmt.Printf("%s events:\n", sensor.sensor.name())

         for _, val := range *sensor.sensor.supported() {
//            if *advanced || !val.advanced {
               fmt.Printf("%30s   %s\n", val.mnemonic, val.desc)
//            }
         }
      }

      os.Exit(0)
   }

   delay := time.Duration(interval) * time.Second
   line := 0

   for {
      time.Sleep(delay)

      // print column headings
      if line == 0 {
         for _, sensor := range sensors {
            for _, event := range sensor.enabled {
               if *discrete {
                  for unit := 0; unit < sensor.units; unit++ {
                     fmt.Printf("%s:%d ", event.mnemonic, unit)
                  }
               } else {
                  fmt.Printf("%s ", event.mnemonic)
               }
            }
         }
         fmt.Println("")
      }
      line = (line + 1) % 25

      for _, sensor := range sensors {
         samples := sensor.sensor.sample()

         for i, event := range sensor.enabled {
            if *discrete {
               for unit := 0; unit < sensor.units; unit++ {
                  fmt.Printf("%*d ", len(event.mnemonic)+2, samples[i*len(sensor.enabled)+unit])
               }
            } else {
               fmt.Printf("%*d ", len(event.mnemonic), samples[i])
            }
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

   // remove any sensors where probe fails
   for i := len(sensors)-1; i >= 0; i-- {
      sensors[i].units = int(sensors[i].sensor.probe())

      if sensors[i].units == 0 {
         sensors = append(sensors[:i], sensors[i+1:]...)
      }
   }

   elems := strings.Split(*events, ",")
   total := 0

   // build a list of enabled events, storing pointer to event struct
   for i := range sensors {
      var enabled []uint16

      for _, elem := range elems {
         supported := *sensors[i].sensor.supported()

         for j := range supported {
            if supported[j].mnemonic == elem {
               enabled = append(enabled, uint16(j))
               sensors[i].enabled = append(sensors[i].enabled, &supported[j])
               total++
            }
         }
      }

      sensors[i].sensor.enable(enabled, *discrete)
   }

   if total == 0 {
      fmt.Println("no matching events")
      os.Exit(0)
   }

   exe := path.Base(os.Args[0])
   if exe == "vmxstat" {
      vmxstat()
      os.Exit(0)
   }

   initweb(*listenAddr)

   for {
      time.Sleep(time.Duration(interval) * time.Second)
//      fmt.Println("update")
   }
}
