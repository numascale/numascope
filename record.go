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
   "encoding/json"
   "os"
   "os/signal"
   "syscall"
   "time"

   "golang.org/x/sys/unix"
)

const (
   fileName = "output.json"
)

var (
   file *os.File
)

func writeLabel(timestamp int64, label string) {
   elems := []interface{}{"label", timestamp, label}
   b, err := json.Marshal(elems)
   validate(err)
   b = append(b, []byte(",\n")...)
   _, err = file.Write(b)
   validate(err)
}

func record() {
   var err error
   file, err = os.Create(fileName)
   validate(err)

   // always capture per-chip counters
   *discrete = true

   fmt.Printf("spooling to %v\n", fileName)

   _, err = file.WriteString("[\n")
   validate(err)

   sigs := make(chan os.Signal, 1)
   signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

   // enable recording all events
   headings := []string{present[0].Name()}
   headings = append(headings, present[0].Headings(false)...)

   present[0].Enable(*discrete)

   b, err := json.Marshal(headings)
   validate(err)
   b = append(b, []byte(",\n")...)
   _, err = file.Write(b)
   validate(err)

   labelBuf := make([]byte, 256)
outer:
   for {
      select {
      case <-sigs:
         break outer
      case <-time.After(time.Duration(interval) * time.Millisecond):
      }

      // record any label
      n, err := unix.Read(fifo, labelBuf)
      validate(err)

      timestamp := time.Now().UnixNano() / 1e3

      if n > 0 {
         writeLabel(timestamp, string(bytes.TrimSpace(labelBuf[:n])))
      }

      line := []int64{timestamp}
      line = append(line, present[0].Sample()...)

      b, err := json.Marshal(line)
      validate(err)
      b = append(b, []byte(",\n")...)
      _, err = file.Write(b)
      validate(err)
   }

   // trim trailing ','
   _, err = file.Seek(-2, os.SEEK_CUR)
   validate(err)

   _, err = file.WriteString("\n]\n")
   validate(err)
}
