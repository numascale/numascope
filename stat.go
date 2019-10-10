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
   "bytes"
   "fmt"
   "os"
   "strings"
   "time"

   "golang.org/x/sys/unix"
)

func stat() {
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

   delay := time.Duration(*interval) * time.Millisecond
   line := 0
   headings := make([][]string, len(present))

   for i, sensor := range present {
      headings[i] = sensor.Headings(true)
   }

   labelBuf := make([]byte, 32)

   for {
      time.Sleep(delay)

      // print any label
      n, err := unix.Read(fifo, labelBuf)
      validateNonblock(err)

      if n > 0 {
         fmt.Printf("- %s -\n", bytes.TrimSpace(labelBuf[:n]))
      }

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
   }
}
