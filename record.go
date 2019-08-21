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
   "fmt"
   "encoding/json"
   "os"
   "os/signal"
   "syscall"
   "time"
)

const (
   fileName = "output.json"
)

func record() {
   f, err := os.Create(fileName)
   validate(err)

   fmt.Printf("spooling to %v\n", fileName)

   _, err = f.WriteString("[\n")
   validate(err)

   sigs := make(chan os.Signal, 1)
   signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

   headings := []string{present[0].Name()}
   for _, sensor := range present {
      headings = append(headings, sensor.Headings()...)
   }

   b, err := json.Marshal(headings)
   validate(err)
   b = append(b, []byte(",\n")...)
   _, err = f.Write(b)
   validate(err)

outer:
   for {
      select {
      case <-sigs:
         break outer
      case <-time.After(time.Duration(interval) * time.Millisecond):
      }

      timestamp := time.Now().UnixNano() / 1e3
      line := []int64{timestamp}

      for _, sensor := range present {
         line = append(line, sensor.Sample()...)
      }

      b, err := json.Marshal(line)
      validate(err)
      b = append(b, []byte(",\n")...)
      _, err = f.Write(b)
      validate(err)
   }

   // trim trailing ','
   _, err = f.Seek(-2, os.SEEK_CUR)
   validate(err)

   _, err = f.WriteString("\n]\n")
   validate(err)
}
