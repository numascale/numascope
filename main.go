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
   interval   = 400
   present    = []Sensor{
      NewNumaconnect2(),
      NewKernel(),
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
      interval *= 1000 // convert to milliseconds
   case flag.NArg() > 1:
      usage()
   }

   if *debug {
      fmt.Printf("detected %v\n", present)
   }

   if *list {
      for _, sensor := range present {
         fmt.Printf("%s events:\n", sensor.Name())

         for _, val := range sensor.Events() {
            fmt.Printf("%30s   %s\n", val.mnemonic, val.desc)
         }
      }

      os.Exit(0)
   }

   delay := time.Duration(interval) * time.Millisecond
   line := 0
   headings := make([][]string, len(present))

   for i, sensor := range present {
      headings[i] = sensor.Headings()
   }

   for {
      time.Sleep(delay)

      // print column headings
      if line == 0 {
         for i := range present {
            fmt.Print(strings.Join(headings[i], " "))
         }
         fmt.Println()
      }

      line = (line + 1) % 25

      for i, sensor := range present {
         samples := sensor.Sample()

         for j, heading := range headings[i] {
            fmt.Printf("%*d ", len(heading), samples[j])
         }
      }
      fmt.Println()

      if flag.NArg() == 0 {
         break
      }
   }
}

func dups() {
   dups := 0

   // check for duplicates
   for _, sensor := range present {
      events := sensor.Events()

      for i := range events {
         for j := range events {
            if i != j && (events[i].mnemonic == events[j].mnemonic || events[i].desc == events[j].desc) {
               fmt.Printf("%s event %d %+v and %d %+v overlap\n", sensor.Name(), i, events[i], j, events[j])
               dups++
            }
         }
      }
   }

   if dups > 0 {
      os.Exit(1)
   }
}

func main() {
   flag.Parse()

   if os.Geteuid() != 0 {
      fmt.Println("please run with sudo/root")
      os.Exit(1)
   }

   // remove any sensors where probe fails
   for i := len(present)-1; i >= 0; i-- {
      if !present[i].Present() {
         present = append(present[:i], present[i+1:]...)
      }
   }

   elems := strings.Split(*events, ",")
   total := 0

   for _, sensor := range present {
      events := sensor.Events()

      for _, elem := range elems {
         for i := range events {
            if events[i].mnemonic == elem {
               events[i].enabled = true
               total++
            }
         }
      }

      sensor.Enable(*discrete)
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
      time.Sleep(time.Duration(interval) * time.Millisecond)
      timestamp := uint64(time.Now().UnixNano() / 1e6)
      var samples []int64

      for _, sensor := range present {
         samples = append(samples, sensor.Sample()...)
      }

      update(timestamp, samples)
   }
}
