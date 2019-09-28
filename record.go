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
   "strconv"
   "strings"
   "syscall"
   "time"

   "golang.org/x/sys/unix"
)

const (
   defaultFilename = "output"
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

func fileStop() {
   if file == nil {
      return
   }

   // trim trailing ','
   _, err := file.Seek(-2, os.SEEK_CUR)
   validate(err)

   _, err = file.WriteString("\n]\n")
   validate(err)

   err = file.Close()
   validate(err)
}

func fileStart(fileName string) {
   fileStop()

   var err error
   fileNameFull := fileName+".json"
   index := 0

again:
   if index > 0 {
      fileNameFull = fmt.Sprintf("%s_%d.json", fileName, index)
   }

   file, err = os.OpenFile(fileNameFull, os.O_CREATE | os.O_EXCL | os.O_WRONLY, 0444)
   if perr, ok := err.(*os.PathError); ok && perr.Err == unix.EEXIST {
      index++
      goto again
   }

   validate(err)

   _, err = file.WriteString("[\n")
   validate(err)

   headings := []string{present[0].Name()}
   headings = append(headings, present[0].Headings(false)...)

   b, err := json.Marshal(headings)
   validate(err)

   b = append(b, []byte(",\n")...)
   _, err = file.Write(b)
   validate(err)

   fmt.Printf("recording to %v\n", fileNameFull)
}

func setInterval(input string) {
   l := len(input)
   if l < 2 {
      fmt.Printf("missing interval")
      return
   }

   suffix := input[l-2:]
   if suffix != "ms" {
      fmt.Printf("unknown suffix '%s'\n", suffix)
      return
   }

   nStr := input[:l-2]

   var err error
   i, err := strconv.Atoi(nStr)
   if err != nil {
      fmt.Printf("unknown number '%s'\n", nStr)
      return
   }

   interval = i
}

func record() {
   // always capture per-chip counters
   *discrete = true
   present[0].Enable(*discrete)

   sigs := make(chan os.Signal, 1)
   signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

   fileStart(defaultFilename)
   fifoBuf := make([]byte, 256)
outer:
   for {
      select {
      case <-sigs:
         break outer
      case <-time.After(time.Duration(interval) * time.Millisecond):
      }

      // handle command
      n, err := unix.Read(fifo, fifoBuf)
      validateNonblock(err)

      timestamp := time.Now().UnixNano() / 1e3

      if n > 0 {
         line := string(bytes.TrimSpace(fifoBuf[:n]))
         fields := strings.SplitN(line, " ", 2)

         switch fields[0] {
         case "record":
            if len(fields) == 2 {
               fileStart(fields[1])
            } else {
               fmt.Println("syntax: record <filename.json>")
            }
         case "label":
            if len(fields) >= 2 {
               writeLabel(timestamp, fields[1])
            } else {
               fmt.Println("syntax: label <label>..")
            }
         case "pause":
            fmt.Printf("pause\n")
         case "resume":
            fmt.Printf("resume\n")
         case "interval":
            if len(fields) == 2 {
               setInterval(fields[1])
            } else {
               fmt.Println("syntax: interval <n>ms")
            }
         default:
            fmt.Printf("unknown command '%v'\n", line)
         }
      }

      line := []int64{timestamp}
      line = append(line, present[0].Sample()...)

      b, err := json.Marshal(line)
      validate(err)
      b = append(b, []byte(",\n")...)
      _, err = file.Write(b)
      validate(err)
   }

   fileStop()
}
