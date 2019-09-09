/*  Copyright (C) 2019 Daniel J Blueman
    This file is part of Numascope.

    Numascope is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    Numascope is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with Numascope.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
   "flag"
   "fmt"
   "os"
   "strings"

   "golang.org/x/sys/unix"
)

const (
   fifoPath = "/run/numascope-ctl"
   coalescing = 600e3
)

var (
// TODO enable advanced when there is useful discrimitation
//   advanced   = flag.Bool("advanced", false, "list all events")
   listenAddr = flag.String("listenAddr", "0.0.0.0:80", "web service listen address and port")
   debug      = flag.Bool("debug", false, "print debugging output")
   events     = flag.String("events", "pgfault,pgalloc_normal,pgfree,numa_local,n2VicBlkXSent,n2RdBlkXSent,n2RdBlkModSent,n2ChangeToDirtySent,n2BcastProbeCmdSent,n2RdRespSent,n2ProbeRespSent", "comma-separated list of events")
   list       = flag.Bool("list", false, "list events available on this host")
   discrete   = flag.Bool("discrete", false, "report events per unit, rather than average")
   interval   = 128

   // highest priority first
   present    = []Sensor{
      NewNumaconnect2(),
      NewKernel(),
   }
   fifo       int
)

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

func pin() {
   var set unix.CPUSet

   for i := 0; i < 4; i++ {
      set.Set(i)
   }

   // attempt, so ignore errors
   unix.SchedSetaffinity(0, &set)
   unix.Setpriority(unix.PRIO_PROCESS, 0, -7)
}

func Activate() {
   for _, sensor := range present {
      sensor.Enable(*discrete)
   }
}

func usage() {
   fmt.Println("Usage: numascope [option...] stat|live|record")
   flag.PrintDefaults()
}

func main() {
   pin()

   flag.Usage = usage
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
   }

   Activate()

   if total == 0 {
      fmt.Println("no matching events")
      os.Exit(0)
   }

   // expected to fail if already exists
   unix.Umask(0)
   unix.Mkfifo(fifoPath, 0666)

   var err error
   fifo, err = unix.Open(fifoPath, unix.O_RDONLY|unix.O_NONBLOCK, 0)
   validate(err)

   if flag.NArg() != 1 {
      flag.Usage()
      os.Exit(1)
   }

   switch flag.Arg(0) {
   case "stat":
      stat()
   case "live":
      live()
   case "record":
      record()
   default:
      // unexpected mode
      flag.Usage()
      os.Exit(1)
   }
}
